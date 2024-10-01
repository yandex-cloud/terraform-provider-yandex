resource "yandex_function_iam_binding" "function-iam" {
  function_id = "your-function-id"
  role        = "serverless.functions.invoker"

  members = [
    "system:allUsers",
  ]
}
