---
subcategory: "Cloud Billing"
page_title: "Yandex: yandex_billing_cloud_binding"
description: |-
  Bind cloud to billing account.
---

# yandex_billing_cloud_binding (Resource)

Bind cloud to billing account. Creating the bind, which connect the cloud to the billing account.
 For more information, see [the official documentation](https://yandex.cloud/docs/billing/operations/pin-cloud).

**Note**: Currently resource deletion do not unbind cloud from billing account. Instead it does no-operations.

## Example usage

```terraform
resource "yandex_billing_cloud_binding" "foo" {
  billing_account_id = "foo-ba-id"
  cloud_id           = "foo-cloud-id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `billing_account_id` (String) The ID of billing account to bind cloud to.
- `cloud_id` (String) Service Instance ID.

### Read-Only

- `id` (String) The resource identifier.

## Import

```bash
# The resource can be imported by using their resource ID.
# For getting a resource ID you can use Yandex Cloud Web UI or YC CLI.

# cloud-binding-id has the following structure - {billing_account_id}/cloud/{cloud_id}`: 
# * {billing_account_id} refers to the billing account id (`foo-ba-id` in example below).
# * {cloud_id}` refers to the cloud id (`foo-cloud-id` in example below). 
# This way `cloud-binding-id` must be equals to `foo-ba-id/cloud/foo-cloud-id`.

# terraform import yandex_billing_cloud_binding.foo cloud-binding-id
terraform import yandex_billing_cloud_binding.foo ...
```
