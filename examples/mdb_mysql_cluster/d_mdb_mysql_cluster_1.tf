//
// Get information about existing MDB MySQL Cluster.
//
data "yandex_mdb_mysql_cluster" "my_cluster" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_mysql_cluster.my_cluster.network_id
}
