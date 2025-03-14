---
subcategory: "Load Testing"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages an Yandex Cloud Load Testing Agent resource.
---

# {{.Name}} ({{.Type}})

A Load Testing Agent resource. For more information, see [the official documentation](https://yandex.cloud/docs/load-testing/concepts/agent).

## Example usage

{{ tffile "examples/loadtesting_agent/r_loadtesting_agent_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the load testing agent. Must be unique within folder.

* `folder_id` - (Required) The ID of the folder that the resources belong to.

* `compute_instance` - (Required) The template for creating new compute instance running load testing agent. The structure is documented below.

* `description` - (Optional) A description of the load testing agent.

* `labels` - (Optional) A set of key/value label pairs to assign to the agent.

* `log_settings` - The logging settings of the load testing agent. The structure is documented below.

---

The `compute_instance` block supports:

* `service_account_id` - (Required) The ID of the service account authorized for this load testing agent. Service account should have `loadtesting.generatorClient` or `loadtesting.externalAgent` role in the folder.

* `resources` - (Required) Compute resource specifications for the instance. The structure is documented below.

* `network_interface` - (Required) Network specifications for the instance. This can be used multiple times for adding multiple interfaces. The structure is documented below.

* `boot_disk` - (Required) Boot disk specifications for the instance. The structure is documented below.

* `zone_id` - (Optional) The availability zone where the virtual machine will be created. If it is not provided, the default provider folder is used.

* `metadata` - (Optional) A set of metadata key/value pairs to make available from within the instance.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance.

* `computed_metadata` - (Computed) The set of metadata `key:value` pairs assigned to this instance. This includes user custom `metadata`, and predefined items created by Yandex Cloud Load Testing.

* `computed_labels` - (Computed) The set of labels `key:value` pairs assigned to this instance. This includes user custom `labels` and predefined items created by Yandex Cloud Load Testing.

* `platform_id` - (Optional) The Compute platform of virtual machine. If it is not provided, the standard-v2 platform will be used.

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

---

The `log_settings` block supports:

* `log_group_id` - The ID of cloud logging group to which the load testing agent sends logs.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 10 minutes
- `delete` - Default 10 minutes
- `update` - Default 10 minutes

 ## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/loadtesting_agent/import.sh" }}
