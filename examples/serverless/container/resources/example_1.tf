resource "yandex_serverless_container_iam_binding" "container-iam" {
  container_id = "your-container-id"
  role         = "serverless.containers.invoker"

  members = [
    "system:allUsers",
  ]
}
