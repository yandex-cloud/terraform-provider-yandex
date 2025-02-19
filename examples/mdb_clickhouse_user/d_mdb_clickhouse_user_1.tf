data "yandex_mdb_clickhouse_user" "foo" {
  cluster_id = "some_cluster_id"
  name     = "username"
  password = "your_password"
}

output "permissions" {
  value = data.yandex_mdb_clickhouse_user.permission
}
