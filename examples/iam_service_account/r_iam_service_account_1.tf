//
// Create a new IAM Service Account (SA).
//
resource "yandex_iam_service_account" "builder" {
  name        = "vmmanager"
  description = "service account to manage VMs"
}
