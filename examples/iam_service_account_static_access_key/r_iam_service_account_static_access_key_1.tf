//
// Create a new IAM Service Account Static Access SKey.
//
resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = "aje5a**********qspd3"
  description        = "static access key for object storage"
  pgp_key            = "keybase:keybaseusername"
}
