resource "yandex_organizationmanager_group_iam_member" "editor" {
  group_id = "some_group_id"
  role     = "editor"
  member   = "userAccount:user_id"
}
