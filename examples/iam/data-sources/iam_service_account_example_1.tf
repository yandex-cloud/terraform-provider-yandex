data "yandex_iam_service_account" "builder" {
  service_account_id = "sa_id"
}

data "yandex_iam_service_account" "deployer" {
  name = "sa_name"
}
