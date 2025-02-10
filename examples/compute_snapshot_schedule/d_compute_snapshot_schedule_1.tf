//
// Get information about existing Compute Snapshot Schedule
//
data "yandex_compute_snapshot_schedule" "my_snapshot_schedule" {
  snapshot_schedule_id = "some_snapshot_schedule_id"
}
