resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = "some_sa_id"
  description        = "static access key for object storage"
  pgp_key            = "keybase:keybaseusername"
}
