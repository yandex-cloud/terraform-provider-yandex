//
// Create a new Compute Snapshot and new IAM Binding for it.
//
resource "yandex_compute_snapshot" "snapshot1" {
  name           = "test-snapshot"
  source_disk_id = "test_disk_id"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_snapshot_iam_binding" "editor" {
  snapshot_id = data.yandex_compute_snapshot.snapshot1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
