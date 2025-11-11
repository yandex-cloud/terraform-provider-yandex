//
// Create a new OrganizationManager Idp Userpool.
//
resource "yandex_organizationmanager_idp_userpool" "example_userpool" {
  name              = "example-userpool"
  organization_id   = "your_organization_id"
  default_subdomain = "example-subdomain"
  description       = "Description example"

  labels = {
    example-label = "example-label-value"
  }

  user_settings = {
    allow_edit_self_login = true
  }
}
