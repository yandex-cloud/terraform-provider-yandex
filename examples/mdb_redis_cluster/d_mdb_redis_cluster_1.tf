//
// Get information about existing MDB Redis Cluster.
//
data "yandex_mdb_redis_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_redis_cluster.foo.network_id
}
