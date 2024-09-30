resource "yandex_compute_snapshot_schedule" "default" {
  name = "my-name"

  schedule_policy {
    expression = "0 0 * * *"
  }

  snapshot_count = 1

  snapshot_spec {
    description = "snapshot-description"
    labels = {
      snapshot-label = "my-snapshot-label-value"
    }
  }

  labels = {
    my-label = "my-label-value"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}

resource "yandex_compute_snapshot_schedule" "default" {
  schedule_policy {
    expression = "0 0 * * *"
  }

  retention_period = "12h"

  snapshot_spec {
    description = "retention-snapshot"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}
