resource "yandex_compute_disk" "default" {
  name     = "disk-name"
  type     = "network-ssd"
  zone     = "ru-central1-a"
  image_id = "ubuntu-16.04-v20180727"

  labels = {
    environment = "test"
  }
}
