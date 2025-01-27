data "yandex_lockbox_secret_version" "my_secret_version" {
  secret_id  = "some-secret-id"
  version_id = "some-version-id" # if you don't indicate it, by default refers to the latest version
}

output "my_secret_entries" {
  value = data.yandex_lockbox_secret_version.my_secret_version.entries
}
