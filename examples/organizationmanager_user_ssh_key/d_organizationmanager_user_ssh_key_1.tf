//
// Get information about existing OrganizationManager User SSH Key.
//
data "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  user_ssh_key_id = "some_user_ssh_key_id"
}

output "my_user_ssh_key_name" {
  value = "data.yandex_organizationmanager_user_ssh_key.my_user_ssh_key.name"
}
