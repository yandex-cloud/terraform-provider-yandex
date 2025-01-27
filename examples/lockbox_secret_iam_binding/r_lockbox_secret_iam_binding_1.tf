resource "yandex_lockbox_secret" "your-secret" {
  name = "secret-name"
}

resource "yandex_lockbox_secret_iam_binding" "viewer" {
  secret_id = yandex_lockbox_secret.your-secret.id
  role      = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
