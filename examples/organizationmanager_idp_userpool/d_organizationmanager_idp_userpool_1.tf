//
// Get information about existing OrganizationManager Idp Userpool.
//
data "yandex_organizationmanager_idp_userpool" "userpool" {
  userpool_id = "some_userpool_id"
}

output "my_userpool.name" {
  value = data.yandex_organizationmanager_idp_userpool.userpool.name
}

output "my_userpool.organization_id" {
  value = data.yandex_organizationmanager_idp_userpool.userpool.organization_id
}

