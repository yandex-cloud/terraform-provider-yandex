//
// Create a new IAM Workload Identity OIDC Federation.
//
resource "yandex_iam_workload_identity_oidc_federation" "wlif" {
  name        = "some_wlif_name"
  folder_id   = "some_folder_id"
  description = "some description"
  disabled    = false
  audiences   = ["aud1", "aud2"]
  issuer      = "https://example-issuer.com"
  jwks_url    = "https://example-issuer.com/jwks"
  labels = {
    key1 = "value1"
    key2 = "value2"
  }
}
