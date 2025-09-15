//
// Create a new Cloud Registry and new IP Permissions for it.
//
resource "yandex_cloudregistry_registry" "my_registry" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
  kind      = "DOCKER"
  type      = "LOCAL"

  description = "Some desctiption"
}

resource "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
  registry_id = yandex_cloudregistry_registry.my_registry.id
  push        = ["10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"]
  pull        = ["10.1.0.0/16", "10.5.0/16"]
}
