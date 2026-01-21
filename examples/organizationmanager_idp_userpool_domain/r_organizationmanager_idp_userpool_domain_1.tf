//
// Create a new OrganizationManager Idp Userpool Domain.
//
resource "yandex_organizationmanager_idp_userpool_domain" "example_domain" {
  userpool_id = "your_userpool_id"
  domain      = "example.com"
}
