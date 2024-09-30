data "yandex_iam_policy" "admin" {
  binding {
    role = "admin"

    members = [
      "userAccount:foobar_user_id",
    ]
  }
}

resource "yandex_iam_service_account_iam_policy" "admin-account-iam" {
  service_account_id = "your-service-account-id"
  policy_data        = data.yandex_iam_policy.admin.policy_data
}
