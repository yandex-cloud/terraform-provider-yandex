//
// Get information about existing MDB Kafka User.
//
data "yandex_mdb_kafka_user" "my_user" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "username" {
  value = data.yandex_mdb_kafka_user.my_user.name
}
