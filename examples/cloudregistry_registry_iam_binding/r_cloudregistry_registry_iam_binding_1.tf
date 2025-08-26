//
// Create a new Cloud Registry and new IAM Binding for it.
//
resource "yandex_cloudregistry_registry" "your-registry" {
  folder_id = "your-folder-id"
  name      = "registry-name"
}

resource "yandex_cloudregistry_registry_iam_binding" "puller" {
  registry_id = yandex_cloudregistry_registry.your-registry.id
  role        = "cloudregistry-registry.images.puller"

  members = [
    "system:allUsers",
  ]
}
