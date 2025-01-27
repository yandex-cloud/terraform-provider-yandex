resource "yandex_compute_placement_group" "group1" {
  name        = "test-pg"
  folder_id   = "abc*********123"
  description = "my description"
}
