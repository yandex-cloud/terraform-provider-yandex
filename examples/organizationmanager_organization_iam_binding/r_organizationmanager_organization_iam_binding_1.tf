//
// Create a new OrganizationManager Organization IAM Binding.
//
resource "yandex_organizationmanager_organization_iam_binding" "editor" {
  organization_id = "some_organization_id"

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
