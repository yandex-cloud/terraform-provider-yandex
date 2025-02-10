---
subcategory: "Cloud Billing"
page_title: "Yandex: {{.Name}}"
description: |-
  Retrieve Yandex Billing cloud to billing account bind details.
---

# {{.Name}} ({{.Type}})

{{ .Description }}

Use this data source to get cloud to billing account bind details. For more information, see [Cloud binding](https://yandex.cloud/docs/billing/operations/pin-cloud).

## Example usage

{{ tffile "examples/billing_cloud_binding/d_billing_cloud_binding_1.tf" }}


## Argument Reference

The following arguments are supported:

* `billing_account_id` - (Required) ID of the billing account.
* `cloud_id` - (Required) ID of the cloud.
