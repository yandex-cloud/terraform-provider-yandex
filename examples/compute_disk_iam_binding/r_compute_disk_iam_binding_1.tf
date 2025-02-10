//
// Create a new Compute Disk and new IAM Binding for it.
//
resource "yandex_compute_disk" "disk1" {
  name     = "disk-name"
  type     = "network-ssd"
  zone     = "ru-central1-a"
  image_id = "ubuntu-16.04-v20180727"

  labels = {
    environment = "test"
  }
}

resource "yandex_compute_disk_iam_binding" "editor" {
  disk_id = data.yandex_compute_disk.disk1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
