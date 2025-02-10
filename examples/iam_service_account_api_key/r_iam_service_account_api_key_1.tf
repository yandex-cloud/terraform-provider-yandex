//
// Create a new IAM Service Account API Key.
//
resource "yandex_iam_service_account_api_key" "sa-api-key" {
  service_account_id = "aje5a**********qspd3"
  description        = "api key for authorization"
  scopes             = ["yc.ydb.topics.manage", "yc.ydb.tables.manage"]
  expires_at         = "2024-11-11T00:00:00Z"
  pgp_key            = "keybase:keybaseusername"
}
