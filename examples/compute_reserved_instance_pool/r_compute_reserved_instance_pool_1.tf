//
// Create a new Compute Reserved Instance Pool
//
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_reserved_instance_pool" "pool" {
  name        = "reserved-instance-pool"
  zone        = "ru-central1-a"
  platform_id = "standard-v2"

  resources_spec = {
    cores         = 4
    core_fraction = 100
    memory        = 4294967296
  }

  boot_disk_spec = {
    image_id = "${data.yandex_compute_image.ubuntu.id}"
  }

  size = 1
}