data "yandex_mdb_postgresql_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = data.yandex_mdb_postgresql_user.foo.permission
}
