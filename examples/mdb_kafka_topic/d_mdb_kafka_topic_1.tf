//
// Get information about existing MDB Kafka Topic.
//
data "yandex_mdb_kafka_topic" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "replication_factor" {
  value = data.yandex_mdb_kafka_topic.foo.replication_factor
}
