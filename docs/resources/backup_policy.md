---
subcategory: "Cloud Backup"
page_title: "Yandex: yandex_backup_policy"
description: |-
  Allows management of Yandex Cloud Backup Policy.
---

# yandex_backup_policy (Resource)

Allows management of [Yandex Cloud Backup Policy](https://yandex.cloud/docs/backup/concepts/policy).

~> Cloud Backup Provider must be activated in order to manipulate with policies. Active it either by UI Console or by `yc` command.

## Example usage

```terraform
resource "yandex_backup_policy" "basic_policy" {
  name = "basic policy"

  scheduling {
    enabled = false
    backup_sets {
      execute_by_interval = 86400
    }
  }

  retention {
    after_backup = false
  }

  reattempts {}

  vm_snapshot_reattempts {}
}
```

For the full policy attributes, take a look at the following example:

```terraform
resource "yandex_backup_policy" "my_policy" {
  archive_name                      = "[Machine Name]-[Plan ID]-[Unique ID]a"
  cbt                               = "USE_IF_ENABLED"
  compression                       = "NORMAL"
  fast_backup_enabled               = true
  format                            = "AUTO"
  multi_volume_snapshotting_enabled = true
  name                              = "example_name"
  performance_window_enabled        = true
  preserve_file_security_settings   = true
  quiesce_snapshotting_enabled      = true
  silent_mode_enabled               = true
  splitting_bytes                   = "9223372036854775807"
  vss_provider                      = "NATIVE"

  reattempts {
    enabled      = true
    interval     = "1m"
    max_attempts = 10
  }

  retention {
    after_backup = false

    rules {
      max_age       = "365d"
      repeat_period = []
    }
  }

  scheduling {
    enabled              = false
    max_parallel_backups = 0
    random_max_delay     = "30m"
    scheme               = "ALWAYS_INCREMENTAL"
    weekly_backup_day    = "MONDAY"


    backup_sets {
      execute_by_time {
        include_last_day_of_month = true
        monthdays                 = []
        months                    = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]
        repeat_at                 = ["04:10"]
        repeat_every              = "30m"
        type                      = "MONTHLY"
        weekdays                  = []
      }
    }
  }

  vm_snapshot_reattempts {
    enabled      = true
    interval     = "1m"
    max_attempts = 10
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` (**Required**) — Name of the policy
- `compression` (Optional. Default: NORMAL) — Archive compression level. Affects CPU. Available values: `"NORMAL"`, `"HIGH"`, `"MAX"`, `"OFF"`.
- `format` (Optional. Default: AUTO) — Format of the backup. It's strongly recommend to leave this option empty or `"AUTO"`. Available values: `"AUTO"`, `"VERSION_11"`, `"VERSION_12"`.
- `multi_volume_snapshotting_enabled` (Optional. Default: true) — If true, snapshots of multiple volumes will be taken simultaneously.
- `preserve_file_security_settings` (Optional. Default: true) — Preserves file security settings. It's better to set this option to true.
- `silent_mode_enabled` (Optional. Default: true) — if true, a user interaction will be avoided when possible.
- `splitting_bytes` (Optional. Default 9223372036854775807) — determines the size to split backups. It's better to leave this option unchanged.
- `vss_provider` (Optional. Default: NATIVE) — Settings for the volume shadow copy service. Available values are: "`NATIVE`", `"TARGET_SYSTEM_DEFINED"`
- `archive_name` (Optional. Default: [Machine Name]-[Plan ID]-[Unique ID]a) — The name of generated archives.
- `performance_window_enabled` (Optional. Default: false) — Time windows for performance limitations of backup.
- `cbt` (Optional. Default: DO_NOT_USE) — Configuration of Changed Block Tracking. Available values are: `"USE_IF_ENABLED"`, `"ENABLED_AND_USE"`, `"DO_NOT_USE"`.
- `quiesce_snapshotting_enabled` (Optional. Default: false) — If true, a quiesced snapshot of the virtual machine will be taken.
- `reattempts` (**Required**) — Amount of reattempts that should be performed while trying to make backup at the host. This attribute consists of the following parameters:
  - `enabled` (Optional. Default: true) — Enable flag
  - `interval` (Optional. Default: "5m") — Retry interval. See `interval_type` for available values
  - `max_attempts` (Optional, Default: 5) — Maximum number of attempts before throwing an error
- `vm_snapshot_reattempts` (Requied) — Amount of reattempts that should be performed while trying to make snapshot. This attribute consists of the following parameters:
  - `enabled` (Optional. Default: true) — Enable flag
  - `interval` (Optional. Default: "5m") — Retry interval. See `interval_type` for available values
  - `max_attempts` (Optional, Default: 5) — Maximum number of attempts before throwing an error
- `retention` (**Required**) — Retention policy for backups. Allows to setup backups lifecycle. This attribute consists of the following parameters:
  - `max_age` (Conflicts with `max_count`) — Deletes backups that older than `max_age`. Exactly one of `max_count` or `max_age` should be set.
  - `max_count` (Conflicts with `max_age`) — Deletes backups if it's count exceeds `max_count`. Exactly one of `max_count` or `max_age` should be set.
  - `after_backup` — Defines whether retention rule applies after creating backup or before.
- `scheduling` (**Required**) — Schedule settings for creating backups on the host.
  - `enabled` (Optional. Default: true) — enables or disables scheduling.
  - `backup_sets` (Required) - A list of schedules with backup sets that compose the whole scheme.
    - `execute_by_interval` (Optional) — Perform backup by interval, since last backup of the host. Maximum value is: 9999 days. See `interval_type` for available values. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.
    - `execute_by_time` (Optional) — Perform backup periodically at specific time. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.
      - `type` (**Required**) — Type of the scheduling. Available values are: `"HOURLY"`, `"DAILY"`, `"WEEKLY"`, `"MONTHLY"`.
      - `weekdays` (Optional. Default: []) — List of weekdays when the backup will be applied. Used in `"WEEKLY"` type.
      - `repeat_at` (Optional. Default: []) — List of time in format `"HH:MM" (24-hours format)`, when the schedule applies.
      - `repeat_every` (Optional) — Frequency of backup repetition. See `interval_type` for available values.
      - `monthdays` (Optional. Default: []) — List of days when schedule applies. Used in `"MONTHLY"` type.
      - `include_last_day_of_month` (Optional. Default: false) — If true, schedule will be applied on the last day of month. See `day_type` for available values.
    - `type` - (Optional. Default: TYPE_AUTO) - BackupSet type. See `backup_set_type` for available values.
  - `max_parallel_backups` (Optional. Default: 0) — Maximum number of backup processes allowed to run in parallel. 0 for unlimited.
  - `random_max_delay` (Optional. Default: 30m) — Configuration of the random delay between the execution of parallel tasks. See `interval_type` for available values.
  - `scheme` (Optional. Default: ALWAYS_INCREMENTAL) — Scheme of the backups. Available values are: `"ALWAYS_INCREMENTAL"`, `"ALWAYS_FULL"`, `"WEEKLY_FULL_DAILY_INCREMENTAL"`, `'WEEKLY_INCREMENTAL"`.
  - `weekly_backup_day` (Optional. Default: MONDAY) — A day of week to start weekly backups. See `day_type` for available values.
  - `execute_by_interval` (Deprecated, use backup_sets instead) — Perform backup by interval, since last backup of the host. Maximum value is: 9999 days. See `interval_type` for available values. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.
  - `execute_by_time` (Deprecated, use backup_sets instead) — Perform backup periodically at specific time. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.
    - `type` (**Required**) — Type of the scheduling. Available values are: `"HOURLY"`, `"DAILY"`, `"WEEKLY"`, `"MONTHLY"`.
    - `weekdays` (Optional. Default: []) — List of weekdays when the backup will be applied. Used in `"WEEKLY"` type.
    - `repeat_at` (Optional. Default: []) — List of time in format `"HH:MM" (24-hours format)`, when the schedule applies.
    - `repeat_every` (Optional) — Frequency of backup repetition. See `interval_type` for available values.
    - `monthdays` (Optional. Default: []) — List of days when schedule applies. Used in `"MONTHLY"` type.
    - `include_last_day_of_month` (Optional. Default: false) — If true, schedule will be applied on the last day of month. See `day_type` for available values.

## Defined types

### interval_type 

A string type, that accepts values in the format of: `number` + `time type`, where `time type` might be:

- `s` — seconds
- `m` — minutes
- `h` — hours
- `d` — days
- `w` — weekdays
- `M` — months

Example of interval value: `"5m", "10d", "2M", "5w"`

### day_type

A string type, that accepts the following values: `"ALWAYS_INCREMENTAL"`, `"ALWAYS_FULL"`, `"WEEKLY_FULL_DAILY_INCREMENTAL"`, `'WEEKLY_INCREMENTAL"`.

### backup_set_type

`"TYPE_AUTO"`, `"TYPE_FULL"`, `"TYPE_INCREMENTAL"`, `'TYPE_DIFFERENTIAL"`.
