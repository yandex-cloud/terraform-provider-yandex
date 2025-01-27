data "yandex_mdb_mongodb_database" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "owner" {
  value = data.yandex_mdb_mongodb_database.foo.name
}
