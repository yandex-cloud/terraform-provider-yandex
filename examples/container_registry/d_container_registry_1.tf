//
// Get information about existing Container Registry.
//
data "yandex_container_registry" "source" {
  registry_id = "some_registry_id"
}
