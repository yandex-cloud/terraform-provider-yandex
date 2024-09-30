resource "yandex_container_registry" "default" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
}

data "yandex_container_registry_ip_permission" "my_ip_permission_by_id" {
  registry_id = yandex_container_registry.default.id
}
