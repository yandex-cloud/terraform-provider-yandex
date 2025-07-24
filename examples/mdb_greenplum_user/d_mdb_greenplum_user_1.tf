//
// Get information about existing MDB Greenplum database user.
//
data "yandex_mdb_greenplum_user" "my_user" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "resource_group" {
  value = data.yandex_mdb_greenplum_user.my_user.resource_group
}
