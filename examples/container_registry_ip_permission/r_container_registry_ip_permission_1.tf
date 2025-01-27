resource "yandex_container_registry" "my_registry" {
  name = "test-registry"
}

resource "yandex_container_registry_ip_permission" "my_ip_permission" {
  registry_id = yandex_container_registry.my_registry.id
  push        = ["10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"]
  pull        = ["10.1.0.0/16", "10.5.0/16"]
}
