resource "yandex_kms_symmetric_key" "example" {
  name        = "example-symetric-key"
  description = "description for key"
}

resource "yandex_kms_secret_ciphertext" "password" {
  key_id      = yandex_kms_symmetric_key.example.id
  aad_context = "additional authenticated data"
  plaintext   = "strong password"
}
