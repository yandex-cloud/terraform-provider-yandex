//
// Get a specific entry from the latest version of a Lockbox secret by key.
//
data "yandex_lockbox_secret_version_entry" "db_password" {
  secret_id = "some-secret-id"
  key       = "db_password"
}

output "db_password" {
  value     = data.yandex_lockbox_secret_version_entry.db_password.text_value
  sensitive = true
}
