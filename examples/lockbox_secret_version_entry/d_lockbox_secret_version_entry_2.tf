//
// Get a specific entry from a pinned version of a Lockbox secret by key.
//
resource "yandex_lockbox_secret" "my_secret" {
  # ...
}

resource "yandex_lockbox_secret_version" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
  entries {
    key        = "db_password"
    text_value = "s3cr3t"
  }
  entries {
    key        = "db_user"
    text_value = "admin"
  }
}

data "yandex_lockbox_secret_version_entry" "db_password" {
  secret_id  = yandex_lockbox_secret.my_secret.id
  version_id = yandex_lockbox_secret_version.my_version.id
  key        = "db_password"
}

output "db_password" {
  value     = data.yandex_lockbox_secret_version_entry.db_password.text_value
  sensitive = true
}
