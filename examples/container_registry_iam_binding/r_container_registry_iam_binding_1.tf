resource "yandex_container_registry" "your-registry" {
  folder_id = "your-folder-id"
  name      = "registry-name"
}

resource "yandex_container_registry_iam_binding" "puller" {
  registry_id = yandex_container_registry.your-registry.id
  role        = "container-registry.images.puller"

  members = [
    "system:allUsers",
  ]
}
