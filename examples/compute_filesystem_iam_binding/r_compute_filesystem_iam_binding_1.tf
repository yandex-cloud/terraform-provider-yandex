resource "yandex_compute_filesystem" "fs1" {
  name = "fs-name"
  type = "network-ssd"
  zone = "ru-central1-a"
  size = 10

  labels = {
    environment = "test"
  }
}

resource "yandex_compute_filesystem_iam_binding" "editor" {
  filesystem_id = data.yandex_compute_filesystem.fs1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
