resource "yandex_iam_service_account" "sa" {
  name        = "vmmanager"
  description = "service account to manage VMs"
}
