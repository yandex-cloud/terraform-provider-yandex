//
// Get information about existing MDB Clickhouse Cluster.
//
data "yandex_mdb_clickhouse_cluster_v2" "my_cluster" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_clickhouse_cluster_v2.my_cluster.network_id
}
