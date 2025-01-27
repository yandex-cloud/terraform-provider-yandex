resource "yandex_kms_asymmetric_encryption_key" "your-key" {
  name      = "asymmetric-encryption-key-name"
}

resource "yandex_kms_asymmetric_encryption_key_iam_binding" "viewer" {
  asymmetric_encryption_key_id = yandex_kms_asymmetric_encryption_key.your-key.id
  role                         = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
