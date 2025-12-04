//
// Create a new OrganizationManager Idp Application SAML Signature Certificate.
//
resource "yandex_organizationmanager_idp_application_saml_signature_certificate" "example_certificate" {
  application_id = "some_application_id"
  name           = "example-signature-certificate"
  description    = "Example signature certificate for SAML application"
}

