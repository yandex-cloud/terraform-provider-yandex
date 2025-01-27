resource "yandex_iam_service_account_key" "sa-auth-key" {
  service_account_id = "some_sa_id"
  description        = "key for service account"
  key_algorithm      = "RSA_4096"
  pgp_key            = "keybase:keybaseusername"
}
