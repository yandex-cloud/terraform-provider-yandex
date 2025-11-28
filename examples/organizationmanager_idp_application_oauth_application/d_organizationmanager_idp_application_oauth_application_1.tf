//
// Get information about existing OrganizationManager Idp Application OAuth Application.
//
data "yandex_organizationmanager_idp_application_oauth_application" "app" {
  application_id = "some_application_id"
}

output "my_app.name" {
  value = data.yandex_organizationmanager_idp_application_oauth_application.app.name
}

output "my_app.status" {
  value = data.yandex_organizationmanager_idp_application_oauth_application.app.status
}

output "my_app.client_id" {
  value = data.yandex_organizationmanager_idp_application_oauth_application.app.client_grant.client_id
}

