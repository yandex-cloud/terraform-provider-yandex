//
// Get information about existing MDB MySQL Database User.
//
data "yandex_mdb_mysql_user" "my_user" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = data.yandex_mdb_mysql_user.foo.permission
}
