//
// Get information about existing OrganizationManager Idp Application SAML Signature Certificate.
//
data "yandex_organizationmanager_idp_application_saml_signature_certificate" "certificate" {
  signature_certificate_id = "some_signature_certificate_id"
}

output "my_certificate.name" {
  value = data.yandex_organizationmanager_idp_application_saml_signature_certificate.certificate.name
}

output "my_certificate.fingerprint" {
  value = data.yandex_organizationmanager_idp_application_saml_signature_certificate.certificate.fingerprint
}

output "my_certificate.status" {
  value = data.yandex_organizationmanager_idp_application_saml_signature_certificate.certificate.status
}

