data "yandex_container_repository" "repo-1" {
  name = "some_repository_name"
}

data "yandex_container_repository" "repo-2" {
  repository_id = "some_repository_id"
}
