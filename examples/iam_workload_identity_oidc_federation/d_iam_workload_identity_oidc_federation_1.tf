data "yandex_iam_workload_identity_oidc_federation" "wlif" {
  federation_id = "some_federation_id"
}

data "yandex_iam_workload_identity_oidc_federation" "wlif" {
  name = "some_federation_name"
}