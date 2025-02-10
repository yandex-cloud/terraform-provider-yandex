//
// Get information about existing MDB Kafka Connector.
//
data "yandex_mdb_kafka_connector" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "tasks_max" {
  value = data.yandex_mdb_kafka_connector.foo.tasks_max
}
