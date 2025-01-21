resource "yandex_compute_placement_group" "pg1" {
  name        = "test-pg"
  folder_id   = "abc*********123"
  description = "my description"
}

resource "yandex_compute_placement_group_iam_binding" "editor" {
  placement_group_id = data.yandex_compute_placement_group.pg1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
