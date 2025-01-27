---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new snapshot of a disk.
---

# {{.Name}} ({{.Type}})

Creates a new snapshot of a disk. For more information, see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/snapshot).

## Example usage

{{ tffile "examples/compute_snapshot/r_compute_snapshot_1.tf" }}

## Argument Reference

The following arguments are supported:

* `source_disk_id` - (Required) ID of the disk to create a snapshot from.

---

* `name` - (Optional) A name for the resource.

* `description` - (Optional) Description of the resource.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the snapshot.

* `hardware_generation` - (Optional) Hardware generation and its features,
  which will be applied to the instance when this snapshot is used as a boot
  disk source. Provide this property if you wish to override this value, which
  otherwise is inherited from the source. The structure is documented below.

The `hardware_generation` consists of one of the following blocks:

* `legacy_features` - Defines the first known hardware generation and its features, which are:
  * `pci_topology` - A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.
* `generation2_features` - A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `disk_size` - Size of the disk when the snapshot was created, specified in GB.
* `storage_size` - Size of the snapshot, specified in GB.
* `created_at` - Creation timestamp of the snapshot.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 20 minutes
- `update` - Default 20 minutes
- `delete` - Default 20 minutes

## Import

A snapshot can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_compute_snapshot.disk-snapshot shapshot_id
```
