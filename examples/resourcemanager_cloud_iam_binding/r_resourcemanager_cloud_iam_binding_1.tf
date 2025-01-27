data "yandex_resourcemanager_cloud" "project1" {
  name = "Project 1"
}

resource "yandex_resourcemanager_cloud_iam_binding" "admin" {
  cloud_id = data.yandex_resourcemanager_cloud.project1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
