//
// Get information about existing YDB Serverless Database.
//
data "yandex_ydb_database_serverless" "my_database" {
  database_id = "some_ydb_serverless_database_id"
}

output "ydb_api_endpoint" {
  value = data.yandex_ydb_database_serverless.my_database.ydb_api_endpoint
}
