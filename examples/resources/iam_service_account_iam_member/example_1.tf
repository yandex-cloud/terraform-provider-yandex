resource "yandex_iam_service_account_iam_member" "admin-account-iam" {
  service_account_id = "your-service-account-id"
  role               = "admin"
  member             = "userAccount:bar_user_id"
}
