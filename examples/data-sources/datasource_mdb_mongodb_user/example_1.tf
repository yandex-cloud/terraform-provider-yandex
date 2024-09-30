data "yandex_mdb_mongodb_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = data.yandex_mdb_mongodb_user.foo.permission
}
