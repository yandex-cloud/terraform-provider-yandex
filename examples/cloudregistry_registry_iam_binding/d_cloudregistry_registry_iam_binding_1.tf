//
// Get information about existing IAM bindings of specific Cloud Registry.
//
data "yandex_cloudregistry_registry_iam_binding" "my_iam_bindings_by_registry_id" {
  registry_id = "some_registry_id"
}
