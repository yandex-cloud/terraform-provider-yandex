//
// Create a new OrganizationManager Group Membership.
//
resource "yandex_organizationmanager_group_membership" "group" {
  group_id = "sdf4*********3fr"
  members = [
    "xdf********123"
  ]
}
