//
// Create a new IAM Binding for existing Folder.
//
data "yandex_resourcemanager_folder" "project1" {
  folder_id = "some_folder_id"
}

resource "yandex_resourcemanager_folder_iam_binding" "admin" {
  folder_id = data.yandex_resourcemanager_folder.project1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
