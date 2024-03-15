---
layout: "yandex"
page_title: "Yandex: yandex_loadtesting_agent"
sidebar_current: "docs-yandex-loadtesting-agent"
description: |-
  Manages an Yandex Cloud Load Testing Agent resource.
---

# yandex\_loadtesting\_agent

A Load Testing Agent resource. For more information, see
[the official documentation](https://cloud.yandex.com/en/docs/load-testing/concepts/agent).

## Example Usage

```hcl
resource "yandex_loadtesting_agent" "my-agent" {
  name = "my-agent"
  description = "2 core 4 GB RAM agent"
  folder_id = "${data.yandex_resourcemanager_folder.test_folder.id}"
  labels = {
    jmeter = "5"
  }
        
  compute_instance {
    zone_id = "ru-central1-b"
    service_account_id = "${yandex_iam_service_account.test_account.id}"
    resources {
        memory = 4
        cores = 2
    }
    boot_disk {
        initialize_params {
            size = 15
        }
        auto_delete = true
    }
    network_interface {
      subnet_id = "${yandex_vpc_subnet.my-subnet-a.id}"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the load testing agent. Must be unique within folder.

* `folder_id` - (Required) The ID of the folder that the resources belong to.

* `compute_instance` - (Required) The template for creating new compute instance running load testing agent. The structure is documented below.

* `description` - (Optional) A description of the load testing agent.

* `labels` - (Optional) A set of key/value label pairs to assign to the agent.

---

The `compute_instance` block supports:

* `service_account_id` - (Required) The ID of the service account authorized for this load testing agent. Service account should have `loadtesting.generatorClient` or `loadtesting.externalAgent` role in the folder.

* `resources` - (Required) Compute resource specifications for the instance. The structure is documented below.

* `network_interface` - (Required) Network specifications for the instance. This can be used multiple times for adding multiple interfaces. The structure is documented below.

* `boot_disk` - (Required) Boot disk specifications for the instance. The structure is documented below.

* `zone_id` - (Optional) The availability zone where the virtual machine will be created. If it is not provided,
    the default provider folder is used.

* `metadata` - (Optional) A set of metadata key/value pairs to make available from within the instance.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance.

* `computed_metadata` - (Computed) The set of metadata `key:value` pairs assigned to this instance. This includes user custom `metadata`, and predefined items created by Yandex Cloud Load Testing.

---

The `resources` block supports:

* `memory` - (Optional) The memory size in GB. Defaults to 2 GB.

* `cores` - (Optional) The number of CPU cores for the instance. Defaults to 2 cores.

* `core_fraction` - (Optional) If provided, specifies baseline core performance as a percent.

---

The `boot_disk` block supports:

* `initialize_params` - (Required) Parameters for creating a disk alongside the instance. The structure is documented below.

* `auto_delete` - (Optional) Whether the disk is auto-deleted when the instance is deleted. The default value is true.

* `device_name` - (Optional) This value can be used to reference the device under `/dev/disk/by-id/`.

* `disk_id` - (Computed) The ID of created disk. 

---

The `initialize_params` block supports:

* `name` - (Optional) A name of the boot disk.

* `description` - (Optional) A description of the boot disk.

* `size` - (Optional) The size of the disk in GB. Defaults to 15 GB.

* `type` - (Optional) The disk type.

* `block_size` - (Optional) Block size of the disk, specified in bytes.

---

The `network_interface` block supports:

* `subnet_id` - (Required) The ID of the subnet to attach this interface to. The subnet must reside in the same zone where this instance was created.

* `ipv4` - (Optional) Flag for allocating IPv4 address for the network interface.

* `ipv6` - (Optional) Flag for allocating IPv6 address for the network interface.

* `nat` - (Optional) Flag for using NAT.

* `nat_ip_address` - (Optional) A public address that can be used to access the internet over NAT.
  
* `security_group_ids` - (Optional) Security group ids for network interface.

* `ip_address` - (Optional) Manual set static IP address.

* `ipv6_address` - (Optional) Manual set static IPv6 address.

## Timeouts

This resource provides the following configuration options for
[timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 30 minutes
- `delete` - Default 30 minutes
