//
// Get information about existing OrganizationManager SAML Federation User by their IAM user account ID.
//
data "yandex_organizationmanager_saml_federation_user" "account" {
  federation_id   = "some_federation_id"
  user_account_id = "some_user_account_id"
}

output "my_federation_user.name_id" {
  value = data.yandex_organizationmanager_saml_federation_user.account.name_id
}

output "my_federation_user.preferred_username" {
  value = data.yandex_organizationmanager_saml_federation_user.account.preferred_username
}
