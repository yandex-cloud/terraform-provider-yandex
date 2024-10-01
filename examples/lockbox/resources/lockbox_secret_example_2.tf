resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret with passowrd"

  password_payload_specification {
    password_key = "some_password"
    length       = 12
  }
}
