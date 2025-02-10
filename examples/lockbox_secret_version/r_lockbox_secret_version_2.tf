//
// Create a new Lockbox Secret Version with password.
//
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret with passowrd"

  password_payload_specification {
    password_key = "some_password"
    length       = 12
  }
}

resource "yandex_lockbox_secret_version" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
}
