//
// Get information about existing Cloud.
//
data "yandex_resourcemanager_cloud" "my_cloud" {
  name = "foo-cloud"
}

output "cloud_create_timestamp" {
  value = data.yandex_resourcemanager_cloud.my_cloud.created_at
}
