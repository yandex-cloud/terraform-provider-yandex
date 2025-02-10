//
// Get information about existing Container Repository 
// and create new IAM Binding for it.
//
data "yandex_container_repository" "repo-2" {
  name = "some_repository_name"
}

resource "yandex_container_repository_iam_binding" "pusher" {
  repository_id = yandex_container_repository.repo-2.id
  role          = "container-registry.images.pusher"

  members = [
    "serviceAccount:your-service-account-id",
  ]
}
