data "yandex_compute_snapshot" "my_snapshot" {
  snapshot_id = "some_snapshot_id"
}

resource "yandex_compute_instance" "default" {
  ...

  boot_disk {
    initialize_params {
      snapshot_id = "${data.yandex_compute_snapshot.my_snapshot.id}"
    }
  }
}
