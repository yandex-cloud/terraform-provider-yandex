resource "yandex_iam_workload_identity_federated_credential" "fc" {
  service_account_id  = "some_sa_id"
  federation_id       = "some_wli_federation_id"
  external_subject_id = "some_external_subject_id"
}