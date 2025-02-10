//
// Get information about existing Compute Filesystem.
//
data "yandex_compute_filesystem" "my_fs" {
  filesystem_id = "some_fs_id"
}

// You can use "data.yandex_compute_filesystem.my_fs.id" identifier 
// as reference to the existing resource.
resource "yandex_compute_instance" "default" {
  # ...

  filesystem {
    filesystem_id = data.yandex_compute_filesystem.my_fs.id
  }
}
