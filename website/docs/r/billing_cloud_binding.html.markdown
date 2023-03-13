---
layout: "yandex"
page_title: "Yandex: yandex_billing_cloud_binding"
sidebar_current: "docs-yandex-billing-cloud-binding"
description: |-
 Bind cloud to billing account.
---

# yandex\_billing\_cloud\_binding

Creating the bind, which connect the cloud to the billing account.
For more information, see [Cloud binding](https://cloud.yandex.ru/docs/billing/operations/pin-cloud).

**Note**: Currently resource deletion do not unbind cloud from billing account. Instead it does no-operations.

## Example Usage

```hcl
resource "yandex_billing_cloud_binding" "foo" {
  billing_account_id = "foo-ba-id"
  cloud_id = "foo-cloud-id"
}
```

## Argument Reference

The following arguments are supported:

* `billing_account_id` - (Required) ID of billing account to bind cloud to.

* `cloud_id` - (Required) ID of cloud to bind.

## Import

Cloud binding can be imported by ID

```
$ terraform import yandex_billing_cloud_binding.foo cloud-binding-id
```

**Note**: `cloud-binding-id` has the following structure `{billing_account_id}/cloud/{cloud_id}`, 
where `{billing_account_id}` refers to the billing account id (`foo-ba-id` in example above) 
and `{cloud_id}` refers to the cloud id (`foo-cloud-id` in example above).
This way `cloud-binding-id` must be equals to `foo-ba-id/cloud/foo-cloud-id`.