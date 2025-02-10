//
// Create a new OrganizationManager User SSH Key.
//
resource "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  organization_id = "some_organization_id"
  subject_id      = "some_subject_id"
  data            = "ssh_key_data"
}
