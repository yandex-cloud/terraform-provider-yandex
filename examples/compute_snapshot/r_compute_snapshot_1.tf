//
// Create a new Compute Snapshot.
//
resource "yandex_compute_snapshot" "default" {
  name           = "test-snapshot"
  source_disk_id = "test_disk_id"

  labels = {
    my-label = "my-label-value"
  }
}
