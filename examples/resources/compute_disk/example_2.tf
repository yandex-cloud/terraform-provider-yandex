resource "yandex_compute_disk" "nr" {
  name = "non-replicated-disk-name"
  size = 93 // NB size must be divisible by 93  
  type = "network-ssd-nonreplicated"
  zone = "ru-central1-b"

  disk_placement_policy {
    disk_placement_group_id = yandex_compute_disk_placement_group.this.id
  }
}

resource "yandex_compute_disk_placement_group" "this" {
  zone = "ru-central1-b"
}
