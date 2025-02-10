//
// Create a new OrganizationManager Group.
//
resource "yandex_organizationmanager_group" "my_group" {
  name            = "my-group"
  description     = "My new Group"
  organization_id = "sdf4*********3fr"
}
