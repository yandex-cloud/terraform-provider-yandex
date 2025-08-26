//
// Create a new Cloud Registry.
//
resource "yandex_cloudregistry_registry" "default" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
}
