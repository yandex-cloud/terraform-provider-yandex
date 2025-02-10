//
// Create a new Lockbox Secret Hashed Version.
//
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret"
}

resource "yandex_lockbox_secret_version_hashed" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
  key_1     = "key1"
  // in Terraform state, these values will be stored in hash format
  text_value_1 = "sensitive value 1"
  key_2        = "k2"
  text_value_2 = "sensitive value 2"
  // etc. (up to 10 entries)
}
