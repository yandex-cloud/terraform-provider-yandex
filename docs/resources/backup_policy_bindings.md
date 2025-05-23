---
subcategory: "Cloud Backup"
page_title: "Yandex: yandex_backup_policy_bindings"
description: |-
  Allows to bind compute instance with backup policy.
---

# yandex_backup_policy_bindings (Resource)

Allows management of [Yandex Cloud Attach and Detach VM](https://yandex.cloud/docs/backup/operations/policy-vm/attach-and-detach-vm).

 ~> Cloud Backup Provider must be activated in order to manipulate with policies.

## Example usage

```terraform
//
// Create a new Cloud Backup Policy Binding
//
resource "yandex_iam_service_account" "test_sa" {
  name = "sa-backup-editor"
}

resource "yandex_resourcemanager_folder_iam_member" "test_binding" {
  folder_id = yandex_iam_service_account.test_sa.folder_id
  role      = "backup.editor"
  member    = "serviceAccount:${yandex_iam_service_account.test_sa.id}"
}

resource "yandex_vpc_network" "test_backup_network" {}

resource "yandex_vpc_subnet" "test_backup_subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test_backup_network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

# https://yandex.cloud/ru/docs/backup/concepts/vm-connection#vm-network-access
resource "yandex_vpc_security_group" "test_backup_security_group" {
  name       = "cloud-backup"
  network_id = yandex_vpc_network.test_backup_network.id
  egress {
    protocol       = "TCP"
    from_port      = 7770
    to_port        = 7800
    v4_cidr_blocks = ["84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 443
    v4_cidr_blocks = ["213.180.204.0/24", "213.180.193.0/24", "178.176.128.0/24", "84.201.181.0/24", "84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 80
    v4_cidr_blocks = ["213.180.204.0/24", "213.180.193.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 8443
    v4_cidr_blocks = ["84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 44445
    v4_cidr_blocks = ["51.250.1.0/24"]
  }
}

data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-2004-lts"
}

resource "yandex_compute_instance" "test_backup_compute" {
  name        = "test-backup-compute"
  platform_id = "standard-v1"
  zone        = "ru-central1-a"

  service_account_id = yandex_iam_service_account.test_sa.id

  network_interface {
    subnet_id          = yandex_vpc_subnet.test_backup_subnet.id
    security_group_ids = [yandex_vpc_security_group.test_backup_security_group.id]
    nat                = true
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
    }
  }

  resources {
    cores  = 2
    memory = 4
  }

  metadata = {
    user-data = "#cloud-config\npackages:\n  - curl\n  - perl\n  - jq\nruncmd:\n  - curl https://storage.yandexcloud.net/backup-distributions/agent_installer.sh | sudo bash\n"
  }
}

data "yandex_backup_policy" "test_backup_policy" {
  name = "Default daily"
}

resource "yandex_backup_policy_bindings" "test_backup_binding" {
  instance_id = yandex_compute_instance.test_backup_compute.id
  policy_id   = data.yandex_backup_policy.test_backup_policy.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `instance_id` (String) Compute Cloud instance ID.
- `policy_id` (String) Backup Policy ID.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `enabled` (Boolean) Flag is specifies whether the policy application is enabled. May be `false` if Processing flag is `true`.
- `id` (String) The ID of this resource.
- `processing` (Boolean) Flag that specifies whether the policy is in the process of binding to an instance.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_backup_policy_bindings.<resource Name> <resource Id>
terraform import yandex_backup_policy_bindings.test_backup_binding ...
```
