//
// Create a new OrganizationManager Idp Application OAuth Application.
//
resource "yandex_organizationmanager_idp_application_oauth_application" "example_app" {
  organization_id = "some_organization_id"
  name           = "example-oauth-app"
  description    = "Example OAuth application"

  client_grant = {
    client_id         = "some_client_id"
    authorized_scopes = ["openid", "profile", "email"]
  }

  group_claims_settings = {
    group_distribution_type = "ALL_GROUPS"
  }

  labels = {
    env = "production"
    app = "example"
  }
}

