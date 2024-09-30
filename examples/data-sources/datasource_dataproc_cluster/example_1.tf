data "yandex_dataproc_cluster" "foo" {
  name = "test"
}

output "service_account_id" {
  value = data.yandex_dataproc_cluster.foo.service_account_id
}
