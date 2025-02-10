//
// Create a new Compute Snapshot Schedule with retention period.
//
resource "yandex_compute_snapshot_schedule" "vm_snap_sch2" {
  schedule_policy {
    expression = "0 0 * * *"
  }

  retention_period = "12h"

  snapshot_spec {
    description = "retention-snapshot"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}
