data "yandex_mdb_elasticsearch_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_elasticsearch_cluster.foo.network_id
}
