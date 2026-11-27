//
// Get information about existing MDB Kafka Cluster.
//
data "yandex_mdb_kafka_cluster" "my_cluster" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_kafka_cluster.my_cluster.network_id
}

output "kraft_controller_zone_ids" {
  value = try(data.yandex_mdb_kafka_cluster.my_cluster.config[0].kraft[0].zone_ids, [])
}
