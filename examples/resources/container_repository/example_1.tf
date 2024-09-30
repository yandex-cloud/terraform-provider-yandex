resource "yandex_container_registry" "my-registry" {
  name = "test-registry"
}

resource "yandex_container_repository" "my-repository" {
  name = "${yandex_container_registry.my-registry.id}/test-repository"
}
