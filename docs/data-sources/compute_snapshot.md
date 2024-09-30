---
subcategory: "Compute"
page_title: "Yandex: yandex_compute_snapshot"
description: |-
  Get information about a Yandex Compute Snapshot.
---


# yandex_compute_snapshot




Get information about a Yandex Compute snapshot. For more information, see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/snapshot).

```terraform
data "yandex_compute_snapshot_schedule" "my_snapshot_schedule" {
  snapshot_schedule_id = "some_snapshot_schedule_id"
}
```

## Argument Reference

The following arguments are supported:

* `snapshot_id` - (Optional) The ID of a specific snapshot.

* `name` - (Optional) The name of the snapshot.

~> **NOTE:** One of `snapshot_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - An optional description of this snapshot.
* `folder_id` - ID of the folder that the snapshot belongs to.
* `storage_size` - The size of the snapshot, specified in Gb.
* `status` - The status of the snapshot.
* `disk_size` - Minimum required size of the disk which is created from this snapshot.
* `source_disk_id` - ID of the source disk.
* `labels` - A map of labels applied to this snapshot.
* `product_ids` - License IDs that indicate which licenses are attached to this snapshot.
* `created_at` - Snapshot creation timestamp.
* `hardware_generation` - Snapshot hardware generation and its features. The structure is documented below.

---

The `hardware_generation` consists of one of the following blocks:

* `legacy_features` - Defines the first known hardware generation and its features, which are:
  * `pci_topology` - A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.
* `generation2_features` - A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.
