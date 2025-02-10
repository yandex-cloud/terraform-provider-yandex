//
// Create new Container Repository and Container Repository Lifecycle Policy for it.
//
resource "yandex_container_registry" "my_registry" {
  name = "test-registry"
}

resource "yandex_container_repository" "my_repository" {
  name = "${yandex_container_registry.my_registry.id}/test-repository"
}

resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
  name          = "test-lifecycle-policy-name"
  status        = "active"
  repository_id = yandex_container_repository.my_repository.id

  rule {
    description  = "my description"
    untagged     = true
    tag_regexp   = ".*"
    retained_top = 1
  }
}
