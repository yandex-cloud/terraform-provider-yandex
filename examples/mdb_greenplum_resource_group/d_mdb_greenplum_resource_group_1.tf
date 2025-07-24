//
// Get information about existing MDB Greenplum database resource group.
//
data "yandex_mdb_greenplum_resource_group" "my_resource_group" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "concurrency" {
  value = data.yandex_mdb_greenplum_resource_group.my_resource_group.concurrency
}
