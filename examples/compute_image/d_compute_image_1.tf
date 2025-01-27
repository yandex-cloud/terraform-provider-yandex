data "yandex_compute_image" "my_image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "default" {
  ...

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.my_image.id}"
    }
  }
  ...
  lifecycle {
    ignore_changes = [boot_disk[0].initialize_params[0].image_id]
  }
}
