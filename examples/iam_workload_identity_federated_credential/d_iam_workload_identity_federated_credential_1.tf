data "yandex_iam_workload_identity_federated_credential" "fc" {
  federated_credential_id = "some_fed_cred_id"
}