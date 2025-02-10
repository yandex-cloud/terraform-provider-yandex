//
// Get information about existing IAM Workload Identity Federated Credential.
//
data "yandex_iam_workload_identity_federated_credential" "fed_cred" {
  federated_credential_id = "some_fed_cred_id"
}
