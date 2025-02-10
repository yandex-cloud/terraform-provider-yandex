//
// Get information about existing IAM Service Account (SA).
//
data "yandex_iam_service_account" "builder" {
  service_account_id = "aje5a**********qspd3"
}

data "yandex_iam_service_account" "deployer" {
  name = "sa_name"
}
