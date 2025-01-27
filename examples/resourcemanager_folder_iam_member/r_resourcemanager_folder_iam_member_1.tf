data "yandex_resourcemanager_folder" "department1" {
  folder_id = "some_folder_id"
}

resource "yandex_resourcemanager_folder_iam_member" "admin" {
  folder_id = data.yandex_resourcemanager.department1.name

  role   = "editor"
  member = "userAccount:user_id"
}
