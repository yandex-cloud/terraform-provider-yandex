//
// Create a new Cloud and new IAM Member for it.
//
data "yandex_resourcemanager_cloud" "department1" {
  name = "Department 1"
}

resource "yandex_resourcemanager_cloud_iam_member" "admin" {
  cloud_id = data.yandex_resourcemanager_cloud.department1.id
  role     = "editor"
  member   = "userAccount:user_id"
}
