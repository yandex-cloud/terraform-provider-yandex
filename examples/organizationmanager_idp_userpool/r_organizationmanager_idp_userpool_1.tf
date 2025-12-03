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

  password_quality_policy = {
    allow_similar   = true
    max_length      = 128
    match_length    = 4
    fixed = {
      lowers_required = true
      uppers_required = true
      digits_required = true
      min_length      = 8
    }
  }
}
