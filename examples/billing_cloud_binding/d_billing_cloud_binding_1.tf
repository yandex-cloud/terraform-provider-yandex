//
// Get information about existing Billing Cloud Binding
//
data "yandex_billing_cloud_binding" "foo" {
  billing_account_id = "foo-ba-id"
  cloud_id           = "foo-cloud-id"
}

output "bound_cloud_id" {
  value = data.yandex_billing_cloud_binding.foo.cloud_id
}
