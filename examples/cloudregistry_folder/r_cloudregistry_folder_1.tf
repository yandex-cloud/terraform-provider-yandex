//
// Create a new Cloud Registry Folder.
//
resource "yandex_cloudregistry_folder" "default" {
  registry_id = "some_registry_id"
  path        = "common-artifacts/some-folder"
}
