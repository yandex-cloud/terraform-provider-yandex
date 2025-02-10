//
// Create a new OrganizationManager SAML Federation.
//
resource "yandex_organizationmanager_saml_federation" "saml_fed" {
  name            = "my-federation"
  description     = "My new SAML federation"
  organization_id = "sdf4*********3fr"
  sso_url         = "https://my-sso.url"
  issuer          = "my-issuer"
  sso_binding     = "POST"
}
