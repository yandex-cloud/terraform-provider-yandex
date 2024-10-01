data "yandex_resourcemanager_cloud" "foo" {
  name = "foo-cloud"
}

output "cloud_create_timestamp" {
  value = data.yandex_resourcemanager_cloud.foo.created_at
}
