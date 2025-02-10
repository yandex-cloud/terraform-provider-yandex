//
// Create a new IAM Service Account IAM Policy.
//
data "yandex_iam_policy" "admin" {
  binding {
    role = "admin"

    members = [
      "userAccount:foobar_user_id",
    ]
  }
}

resource "yandex_iam_service_account_iam_policy" "admin-account-iam" {
  service_account_id = "aje5a**********qspd3"
  policy_data        = data.yandex_iam_policy.admin.policy_data
}
