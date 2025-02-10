//
// Create a new Cloud Function IAM Binding.
//
resource "yandex_function_iam_binding" "function-iam" {
  function_id = "dns9m**********tducf"
  role        = "serverless.functions.invoker"

  members = [
    "system:allUsers",
  ]
}
