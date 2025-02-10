//
// Create a new Compute Snapshot Schedule and new IAM Binding for it.
//
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
