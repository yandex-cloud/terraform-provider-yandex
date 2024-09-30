data "yandex_backup_policy" "my_policy" {
  name = "some_policy_name"
}

output "my_policy_name" {
  value = data.yandex_backup_policy.my_policy.name
}
