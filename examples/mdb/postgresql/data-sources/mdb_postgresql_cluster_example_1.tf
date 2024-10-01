data "yandex_mdb_postgresql_cluster" "foo" {
  name = "test"
}

output "fqdn" {
  value = data.yandex_mdb_postgresql_cluster.foo.host.0.fqdn
}
