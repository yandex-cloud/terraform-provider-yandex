//
// Create a new Lockbox Secret and new IAM Member for it.
//
resource "yandex_lockbox_secret" "your-secret" {
  name = "secret-name"
}

resource "yandex_lockbox_secret_iam_member" "viewer" {
  secret_id = yandex_lockbox_secret.your-secret.id
  role      = "viewer"

  member = "userAccount:foo_user_id"
}
