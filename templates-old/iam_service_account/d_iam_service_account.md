---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex IAM service account.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex IAM service account. For more information about accounts, see [Yandex Cloud IAM accounts](https://yandex.cloud/docs/iam/concepts/#accounts).

## Example usage

{{ tffile "examples/iam_service_account/d_iam_service_account_1.tf" }}

## Argument reference

* `service_account_id` - (Optional) ID of a specific service account.

* `name` - (Optional) Name of a specific service account.

~> One of `service_account_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

* `description` - Description of the service account.
