resource yandex_container_registry your-registry {
  folder_id = "your-folder-id"
  name      = "registry-name"
}

resource yandex_container_repository repo-1 {
  name      = "${yandex_container_registry.your-registry.id}/repo-1"
}

resource "yandex_container_repository_iam_binding" "puller" {
  repository_id = yandex_container_repository.repo-1.id
  role        = "container-registry.images.puller"

  members = [
    "system:allUsers",
  ]
}

data "yandex_container_repository" "repo-2" {
  name = "some_repository_name"
}

resource "yandex_container_repository_iam_binding" "pusher" {
  repository_id = yandex_container_repository.repo-2.id
  role        = "container-registry.images.pusher"

  members = [
    "serviceAccount:your-service-account-id",
  ]
}
