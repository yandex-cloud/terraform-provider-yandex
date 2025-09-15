//
// Create a new Cloud Registry and new IAM Binding for it.
//
resource "yandex_cloudregistry_registry" "your-registry" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
  kind      = "DOCKER"
  type      = "LOCAL"

  description = "Some desctiption"
}

resource "yandex_cloudregistry_registry_iam_binding" "puller" {
  registry_id = yandex_cloudregistry_registry.your-registry.id
  role        = "cloud-registry.artifacts.puller"

  members = [
    "system:allUsers",
  ]
}
