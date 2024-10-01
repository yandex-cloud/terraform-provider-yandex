resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name                 = "example-asymetric-encryption-key"
  description          = "description for key"
  encryption_algorithm = "RSA_2048_ENC_OAEP_SHA_256"
}
