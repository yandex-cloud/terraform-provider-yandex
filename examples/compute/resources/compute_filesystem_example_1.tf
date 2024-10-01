resource "yandex_compute_filesystem" "default" {
  name = "fs-name"
  type = "network-ssd"
  zone = "ru-central1-a"
  size = 10

  labels = {
    environment = "test"
  }
}
