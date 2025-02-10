//
// Get information about existing MDB SQL Server Cluster.
//
data "yandex_mdb_sqlserver_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_sqlserver_cluster.foo.network_id
}
