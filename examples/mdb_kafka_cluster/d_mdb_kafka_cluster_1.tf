//
// Get information about existing MDB Kafka Cluster.
//
data "yandex_mdb_kafka_cluster" "my_cluster" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_kafka_cluster.my_cluster.network_id
}
