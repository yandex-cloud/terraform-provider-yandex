---
layout: "yandex"
page_title: "Yandex: yandex_billing_cloud_binding"
sidebar_current: "docs-yandex-datasource-billing-cloud-binding"
description: |-
  Retrieve Yandex Billing cloud to billing account bind details.
---

# yandex\_billing\_cloud\_binding

Use this data source to get cloud to billing account bind details.
For more information, see [Cloud binding](https://cloud.yandex.ru/docs/billing/operations/pin-cloud).

## Example Usage

```hcl
data "yandex_billing_cloud_binding" "foo" {
  billing_account_id = "foo-ba-id"
  cloud_id = "foo-cloud-id"
}

output "bound_cloud_id" {
  value = "${data.yandex_billing_cloud_binding.foo.cloud_id}"
}
```

## Argument Reference

The following arguments are supported:

* `billing_account_id` - (Required) ID of the billing account.
* `cloud_id` - (Required) ID of the cloud.