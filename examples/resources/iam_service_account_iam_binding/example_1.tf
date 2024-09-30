resource "yandex_iam_service_account_iam_binding" "admin-account-iam" {
  service_account_id = "your-service-account-id"
  role               = "admin"

  members = [
    "userAccount:foo_user_id",
  ]
}
