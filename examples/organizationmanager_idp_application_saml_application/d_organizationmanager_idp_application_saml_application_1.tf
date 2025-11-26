//
// Get information about existing OrganizationManager Idp SAML Application.
//
data "yandex_organizationmanager_idp_application_saml_application" "saml_app" {
  application_id = "some_application_id"
}

output "my_saml_app.name" {
  value = data.yandex_organizationmanager_idp_application_saml_application.saml_app.name
}

output "my_saml_app.organization_id" {
  value = data.yandex_organizationmanager_idp_application_saml_application.saml_app.organization_id
}

output "my_saml_app.status" {
  value = data.yandex_organizationmanager_idp_application_saml_application.saml_app.status
}

