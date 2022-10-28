---
layout: "yandex"
page_title: "Yandex: yandex_compute_snapshot_schedule"
sidebar_current: "docs-yandex-datasource-compute-snapshot-schedule"
description: |-
  Get information about a Yandex Compute SnapshotSchedule.
---

# yandex\_compute\_snapshot\_schedule

Get information about a Yandex Compute snapshot schedule. For more information, see
[the official documentation](https://cloud.yandex.ru/docs/compute/concepts/snapshot-schedule).

## Example Usage

```hcl
data "yandex_compute_snapshot_schedule" "my_snapshot_schedule" {
  snapshot_schedule_id = "some_snapshot_schedule_id"
}
```

## Argument Reference

The following arguments are supported:

* `snapshot_schedule_id` - (Optional) The ID of a specific snapshot schedule.

* `name` - (Optional) The name of the snapshot schedule.

~> **NOTE:** One of `snapshot_schedule_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `created_at` - SnapshotSchedule creation timestamp.
* `description` - An optional description of this snapshot schedule.
* `folder_id` - ID of the folder that the snapshot schedule belongs to.
* `labels` - A map of labels applied to this snapshot schedule.
* `retention_period` - Retention period applied to snapshots created by this snapshot schedule.
* `schedule_policy` - Schedule policy of the snapshot schedule.
* `snapshot_count` - Maximum number of snapshots for every disk of the snapshot schedule.
* `snapshot_spec` - Additional attributes for snapshots created by this snapshot schedule.
* `status` - The status of the snapshot schedule.
* `disk_ids` - IDs of the disks of this snapshot schedule.
