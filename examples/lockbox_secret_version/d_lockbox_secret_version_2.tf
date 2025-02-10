//
// Get information about existing Lockbox Secret Version.
//
resource "yandex_lockbox_secret" "my_secret" {
  # ...
}

resource "yandex_lockbox_secret_version" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
  # ...
}

data "yandex_lockbox_secret_version" "my_version" {
  secret_id  = yandex_lockbox_secret.my_secret.id
  version_id = yandex_lockbox_secret_version.my_version.id
}

output "my_secret_entries" {
  value = data.yandex_lockbox_secret_version.my_version.entries
}
