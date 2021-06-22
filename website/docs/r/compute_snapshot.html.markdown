---
layout: "yandex"
page_title: "Yandex: yandex_compute_snapshot"
sidebar_current: "docs-yandex-compute-snapshot"
description: |-
  Creates a new snapshot of a disk.
---

# yandex\_compute\_snapshot

Creates a new snapshot of a disk. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/snapshot).

## Example Usage

```hcl
resource "yandex_compute_snapshot" "default" {
  name           = "test-snapshot"
  source_disk_id = "test_disk_id"

  labels = {
    my-label = "my-label-value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `source_disk_id` - (Required) ID of the disk to create a snapshot from.

- - -

* `name` - (Optional) A name for the resource.

* `description` - (Optional) Description of the resource.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the snapshot.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `disk_size` - Size of the disk when the snapshot was created, specified in GB.
* `storage_size` - Size of the snapshot, specified in GB.
* `created_at` - Creation timestamp of the snapshot.

## Timeouts

This resource provides the following configuration options for
[timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 20 minutes
- `update` - Default 20 minutes
- `delete` - Default 20 minutes

## Import

A snapshot can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_compute_snapshot.disk-snapshot shapshot_id
```
