//
// Get information about existing MDB PostgreSQL Cluster.
//
data "yandex_mdb_postgresql_cluster" "my_cluster" {
  name = "test"
}

output "fqdn" {
  value = data.yandex_mdb_postgresql_cluster.my_cluster.host.0.fqdn
}
