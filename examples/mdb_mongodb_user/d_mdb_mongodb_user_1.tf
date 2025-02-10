//
// Get information about existing MDB MongoDB User.
//
data "yandex_mdb_mongodb_user" "my_user" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = data.yandex_mdb_mongodb_user.my_user.permission
}
