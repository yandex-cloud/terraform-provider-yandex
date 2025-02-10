//
// Get information about existing Compute Disk.
//
data "yandex_compute_disk" "my_disk" {
  disk_id = "some_disk_id"
}

// You can use "data.yandex_compute_disk.my_disk.id" identifier 
// as reference to the existing resource.
resource "yandex_compute_instance" "default" {
  # ...

  secondary_disk {
    disk_id = data.yandex_compute_disk.my_disk.id
  }
}
