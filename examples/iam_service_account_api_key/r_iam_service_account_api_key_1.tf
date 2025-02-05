resource "yandex_iam_service_account_api_key" "sa-api-key" {
  service_account_id = "some_sa_id"
  description        = "api key for authorization"
  scopes             = ["yc.ydb.topics.manage", "yc.ydb.tables.manage"]
  expires_at         = "2024-11-11T00:00:00Z"
  pgp_key            = "keybase:keybaseusername"
}
