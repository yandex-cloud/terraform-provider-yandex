data "yandex_mdb_greenplum_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_greenplum_cluster.foo.network_id
}
