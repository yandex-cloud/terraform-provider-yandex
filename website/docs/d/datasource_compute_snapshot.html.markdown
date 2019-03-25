---
layout: "yandex"
page_title: "Yandex: yandex_compute_snapshot"
sidebar_current: "docs-yandex-datasource-compute-snapshot"
description: |-
  Get information about a Yandex Compute Snapshot.
---

# yandex\_compute\_snapshot

Get information about a Yandex Compute snapshot. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/snapshot).

## Example Usage

```hcl
data "yandex_compute_snapshot" "my_snapshot" {
  snapshot_id = "some_snapshot_id"
}

resource "yandex_compute_instance" "default" {
  ...

  boot_disk {
    initialize_params {
      snapshot_id = "${data.yandex_compute_snapshot.my_snapshot.id}"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `snapshot_id` - (Optional) The ID of a specific snapshot.

* `name` - (Optional) The name of the snapshot.

~> **NOTE:** One of `snapshot_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `description` - An optional description of this snapshot.
* `folder_id` - ID of the folder that the snapshot belongs to.
* `storage_size` - The size of the snapshot, specified in Gb.
* `status` - The status of the snapshot.
* `disk_size` - Minimum required size of the disk which is created from this snapshot.
* `source_disk_id` - ID of the source disk.
* `labels` - A map of labels applied to this snapshot.
* `product_ids` - License IDs that indicate which licenses are attached to this snapshot.
* `created_at` - Snapshot creation timestamp.
