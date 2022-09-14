---
layout: "yandex"
page_title: "Yandex: yandex_ydb_database_dedicated"
sidebar_current: "docs-yandex-ydb-database-dedicated"
description: |-
  Manages Yandex Database dedicated cluster.
---

# yandex\_ydb\_database\_dedicated

Yandex Database (dedicated) resource.
For more information, see [the official documentation](https://cloud.yandex.com/en/docs/ydb/concepts/serverless_and_dedicated).

## Example Usage

```hcl
resource "yandex_ydb_database_dedicated" "database1" {
  name      = "test-ydb-dedicated"
  folder_id = "${data.yandex_resourcemanager_folder.test_folder.id}"

  network_id = "${yandex_vpc_network.my-inst-group-network.id}"
  subnet_ids = ["${yandex_vpc_subnet.my-inst-group-subnet.id}"]

  resource_preset_id  = "medium"
  deletion_protection = true

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  location {
    region {
      id = "ru-central1"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Yandex Database cluster.

* `network_id` - (Required) ID of the network to attach the Yandex Database cluster to.

* `subnet_ids` - (Required) List of subnet IDs to attach the Yandex Database cluster to.

* `resource_preset_id` - (Required) The Yandex Database cluster preset.
  Available presets can be obtained via `yc ydb resource-preset list` command.

* `scale_policy` - (Required) Scaling policy for the Yandex Database cluster.
  The structure is documented below.

* `storage_config` - (Required) A list of storage configuration options for the Yandex Database cluster.
  The structure is documented below.

* `location` - (Optional) Location for the Yandex Database cluster.
  The structure is documented below.

* `location_id` - (Optional) Location ID for the Yandex Database cluster.

* `assign_public_ips` - (Optional) Whether public IP addresses should be assigned to the Yandex Database cluster.

* `folder_id` - (Optional) ID of the folder that the Yandex Database cluster belongs to.
  It will be deduced from provider configuration if not set explicitly.

* `description` - (Optional) A description for the Yandex Database cluster.

* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Database cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the database. Can be either `true` or `false`

---

The `scale_policy` block supports:

* `fixed_scale` - (Required) Fixed scaling policy for the Yandex Database cluster.
  The structure is documented below.

~> **NOTE:** Currently, only `fixed_scale` is supported.

---

The `fixed_scale` block supports:

* `size` - (Required) Number of instances for the Yandex Database cluster.

---

The `storage_config` block supports:

* `storage_type_id` - (Required) Storage type ID for the Yandex Database cluster.
  Available presets can be obtained via `yc ydb storage-type list` command.

* `group_count` - (Required) Amount of storage groups of selected type for the Yandex Database cluster.

---

The `location` block supports:

* `region` - (Optional) Region for the Yandex Database cluster.
  The structure is documented below.

---

The `region` block supports:

* `id` - (Required) Region ID for the Yandex Database cluster.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the Yandex Database cluster.

* `ydb_full_endpoint` - Full endpoint of the Yandex Database cluster.

* `ydb_api_endpoint` - API endpoint of the Yandex Database cluster.
  Useful for SDK configuration.

* `database_path` - Full database path of the Yandex Database cluster.
  Useful for SDK configuration.

* `tls_enabled` - Whether TLS is enabled for the Yandex Database cluster.
  Useful for SDK configuration.

* `created_at` - The Yandex Database cluster creation timestamp.

* `status` - Status of the Yandex Database cluster.
