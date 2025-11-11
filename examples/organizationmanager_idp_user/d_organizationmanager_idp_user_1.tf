//
// Get information about existing OrganizationManager Idp User.
//
data "yandex_organizationmanager_idp_user" "user" {
  user_id = "some_user_id"
}

output "my_user.username" {
  value = data.yandex_organizationmanager_idp_user.user.username
}

output "my_user.full_name" {
  value = data.yandex_organizationmanager_idp_user.user.full_name
}

