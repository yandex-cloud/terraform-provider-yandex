//
// Get information about existing MDB PostgreSQL Database.
//
data "yandex_mdb_postgresql_database" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "owner" {
  value = data.yandex_mdb_postgresql_database.foo.owner
}
