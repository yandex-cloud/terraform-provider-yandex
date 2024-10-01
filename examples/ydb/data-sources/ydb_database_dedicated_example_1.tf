data "yandex_ydb_database_dedicated" "my_database" {
  database_id = "some_ydb_dedicated_database_id"
}

output "ydb_api_endpoint" {
  value = data.yandex_ydb_database_dedicated.my_database.ydb_api_endpoint
}
