data "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy_by_id" {
  lifecycle_policy_id = yandex_container_repository_lifecycle_policy.my_lifecycle_policy.id
}
