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
