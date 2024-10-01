resource "yandex_kms_asymmetric_signature_key" "key-a" {
  name                = "example-asymetric-signature-key"
  description         = "description for key"
  signature_algorithm = "RSA_2048_SIGN_PSS_SHA_256"
}
