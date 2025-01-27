# The resource can be imported by using their resource ID.
# For getting a resource ID you can use Yandex Cloud Web UI or YC CLI.

# cloud-binding-id has the following structure - {billing_account_id}/cloud/{cloud_id}`: 
# * {billing_account_id} refers to the billing account id (`foo-ba-id` in example below).
# * {cloud_id}` refers to the cloud id (`foo-cloud-id` in example below). 
# This way `cloud-binding-id` must be equals to `foo-ba-id/cloud/foo-cloud-id`.

# terraform import yandex_billing_cloud_binding.foo cloud-binding-id
terraform import yandex_billing_cloud_binding.foo ...
