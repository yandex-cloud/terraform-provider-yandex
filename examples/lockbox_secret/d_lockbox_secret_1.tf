data "yandex_lockbox_secret" "my_secret" {
  secret_id = "some ID"
}

output "my_secret_created_at" {
  value = data.yandex_lockbox_secret.my_secret.created_at
}
