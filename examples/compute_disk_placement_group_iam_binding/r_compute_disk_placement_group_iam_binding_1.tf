//
// Create a new Disk Placement Group and new IAM Binding for it.
//
resource "yandex_compute_disk_placement_group" "group1" {
  name        = "test-pg"
  folder_id   = "abc*********123"
  description = "my description"
}

resource "yandex_compute_disk_placement_group_iam_binding" "editor" {
  disk_placement_group_id = data.yandex_compute_disk_placement_group.group1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
