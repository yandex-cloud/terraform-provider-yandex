data "yandex_organizationmanager_saml_federation" "federation" {
  federation_id   = "some_federation_id"
  organization_id = "some_organization_id"
}

output "my_federation.name" {
  value = data.yandex_organizationmanager_saml_federation.federation.name
}
