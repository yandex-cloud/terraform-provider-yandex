//
// Create a new IAM Policy for existing Folder.
//
data "yandex_resourcemanager_folder" "project1" {
  folder_id = "my_folder_id"
}

data "yandex_iam_policy" "admin" {
  binding {
    role = "editor"

    members = [
      "userAccount:some_user_id",
    ]
  }
}

resource "yandex_resourcemanager_folder_iam_policy" "folder_admin_policy" {
  folder_id   = data.yandex_folder.project1.id
  policy_data = data.yandex_iam_policy.admin.policy_data
}
