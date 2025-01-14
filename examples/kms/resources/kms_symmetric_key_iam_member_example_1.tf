resource "yandex_kms_symmetric_key" "your-key" {
  name      = "symmetric-key-name"
}

resource "yandex_kms_symmetric_key_iam_member" "viewer" {
  symmetric_key_id = yandex_kms_symmetric_key.your-key.id
  role             = "viewer"

  member = "userAccount:foo_user_id"
}
