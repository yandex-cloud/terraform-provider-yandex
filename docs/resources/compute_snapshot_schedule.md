---
subcategory: "Compute"
page_title: "Yandex: yandex_compute_snapshot_schedule"
description: |-
  Creates a new snapshot schedule.
---


# yandex_compute_snapshot_schedule




Creates a new snapshot schedule. For more information, see [the official documentation](https://cloud.yandex.ru/docs/compute/concepts/snapshot-schedule).

```terraform
resource "yandex_compute_snapshot_schedule" "schedule1" {
  schedule_policy {
    expression = "0 0 * * *"
  }

  retention_period = "12h"

  snapshot_spec {
    description = "retention-snapshot"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}

resource "yandex_compute_snapshot_schedule_iam_binding" "editor" {
  snapshot_schedule_id = data.yandex_compute_snapshot_schedule.schedule1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `schedule_policy` - (Required) Schedule policy of the snapshot schedule.
* `disk_ids` - (Optional) IDs of the disk for snapshot schedule.
* `retention_period` - (Optional) Time duration applied to snapshots created by this snapshot schedule. This is a signed sequence of decimal numbers, each with optional fraction and a unit suffix. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". Examples: "300ms", "1.5h" or "2h45m".
* `snapshot_count` - (Optional) Maximum number of snapshots for every disk of the snapshot schedule.
* `snapshot_spec` - (Optional) Additional attributes for snapshots created by this snapshot schedule.

---

* `name` - (Optional) A name for the resource.

* `description` - (Optional) Description of the resource.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the snapshot schedule.

The `snapshot_spec` block supports:

* `description` - (Optional) Description to assign to snapshots created by this snapshot schedule.

* `labels` - (Optional) A set of key/value label pairs to assign to snapshots created by this snapshot schedule.

The `schedule_policy` block supports:

* `expression` - (Required) Cron expression to schedule snapshots (in cron format "* * * * *").

* `start_at` - (Optional) Time to start the snapshot schedule (in format RFC3339 "2006-01-02T15:04:05Z07:00"). If empty current time will be used. Unlike an `expression` that specifies regularity rules, the `start_at` parameter determines from what point these rules will be applied.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the snapshot schedule.
* `status` - The status of the snapshot schedule.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 5 minutes
- `update` - Default 5 minutes
- `delete` - Default 5 minutes

## Import

A snapshot schedule can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_compute_snapshot_schedule.my-schedule snapshot_schedule_id
```
