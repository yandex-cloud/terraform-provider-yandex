//
// Create a new Compute Disk and put it to the specific Placement Group.
//
resource "yandex_compute_disk" "my_vm" {
  name = "non-replicated-disk-name"
  size = 93 // Non-replicated SSD disk size must be divisible by 93G
  type = "network-ssd-nonreplicated"
  zone = "ru-central1-b"

  disk_placement_policy {
    disk_placement_group_id = yandex_compute_disk_placement_group.my_pg.id
  }
}

resource "yandex_compute_disk_placement_group" "my_pg" {
  zone = "ru-central1-b"
}
