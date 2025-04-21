//
// Create a new IAM workload identity federation IAM Binding.
//
resource "yandex_iam_workload_identity_oidc_federation_iam_binding" "viewer" {
  federation_id = "example_federation_id"
  role          = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
