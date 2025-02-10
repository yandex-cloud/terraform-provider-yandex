//
// Create a new full Cloud Backup Policy
//
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
