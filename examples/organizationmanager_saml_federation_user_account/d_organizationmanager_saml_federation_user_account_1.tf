data "yandex_organizationmanager_saml_federation_user_account" "account" {
  federation_id = "some_federation_id"
  name_id       = "example@example.org"
}

output "my_federation.id" {
  value = data.yandex_organizationmanager_saml_federation_user_account.account.id
}
