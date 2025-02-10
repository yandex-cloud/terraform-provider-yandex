//
// Create a new OrganizationManager Group Mapping.
//
resource "yandex_organizationmanager_group_mapping" "my_group_map" {
  federation_id = "my-federation-id"
  enabled       = true
}
