data "yandex_mdb_clickhouse_database" "foo" {
  cluster_id = "some_cluster_id"
  name     = "dbname"
}

output "dbname" {
  value = data.yandex_mdb_clickhouse_database.foo.id
}
