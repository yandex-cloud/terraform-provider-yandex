//
// Get information about existing Cloud Registry.
//
data "yandex_cloudregistry_registry" "source" {
  registry_id = "some_registry_id"
}
