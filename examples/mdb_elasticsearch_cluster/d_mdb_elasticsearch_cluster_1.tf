//
// Get information about existing MDB ElasticSearch Cluster.
//
data "yandex_mdb_elasticsearch_cluster" "my_cluster" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_elasticsearch_cluster.my_cluster.network_id
}
