//
// Create a new OrganizationManager Idp SAML Application.
//
resource "yandex_organizationmanager_idp_application_saml_application" "example_saml_app" {
  name            = "example-saml-app"
  organization_id = "your_organization_id"
  description     = "Example SAML application"

  service_provider = {
    entity_id = "https://example.com/saml/metadata"

    acs_urls = [
      {
        url = "https://example.com/saml/acs"
      }
    ]

    slo_urls = [
      {
        url              = "https://example.com/saml/slo"
        protocol_binding = "HTTP_POST"
      }
    ]
  }

  attribute_mapping = {
    name_id = {
      format = "EMAIL"
    }

    attributes = [{
      name  = "email"
      value = "SubjectClaims.email"
    }, {
      name  = "firstName"
      value = "SubjectClaims.given_name"
    }, {
      name  = "lastName"
      value = "SubjectClaims.family_name"
    }]
  }

  security_settings = {
    signature_mode = "RESPONSE_AND_ASSERTIONS"
  }

  labels = {
    environment = "production"
    app-type    = "saml"
  }
}

