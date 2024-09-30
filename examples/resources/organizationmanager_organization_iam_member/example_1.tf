resource "yandex_organizationmanager_organization_iam_member" "editor" {
  organization_id = "some_organization_id"
  role            = "editor"
  member          = "userAccount:user_id"
}
