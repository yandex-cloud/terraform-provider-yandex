//
// Get information about existing MDB Redis User.
//
data "yandex_mdb_redis_user" "my_user" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permissions" {
  value = data.yandex_mdb_redis_user.my_user.permissions
}