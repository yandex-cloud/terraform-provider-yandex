data "yandex_mdb_kafka_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
  password   = "pass123"
}

output "username" {
  value = data.yandex_mdb_kafka_user.foo.name
}
