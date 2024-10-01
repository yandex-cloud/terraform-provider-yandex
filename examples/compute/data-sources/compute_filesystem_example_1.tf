data "yandex_compute_filesystem" "my_fs" {
  filesystem_id = "some_fs_id"
}

resource "yandex_compute_instance" "default" {
  ...

  filesystem {
    filesystem_id = "${data.yandex_compute_filesystem.my_fs.id}"
  }
}
