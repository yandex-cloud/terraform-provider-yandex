//
// Create a new IAM Service Account IAM Binding.
//
resource "yandex_iam_service_account_iam_binding" "admin-account-iam" {
  service_account_id = "aje5a**********qspd3"
  role               = "admin"

  members = [
    "userAccount:foo_user_id",
  ]
}
