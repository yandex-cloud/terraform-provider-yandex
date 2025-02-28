data "yandex_mdb_redis_cluster_v2" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_redis_cluster_v2.foo.network_id
}
