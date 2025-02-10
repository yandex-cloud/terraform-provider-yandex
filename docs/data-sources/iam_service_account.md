---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_service_account"
description: |-
  Get information about a Yandex IAM service account.
---

# yandex_iam_service_account (Data Source)

Get information about a Yandex IAM service account. For more information about accounts, see [Yandex Cloud IAM accounts](https://yandex.cloud/docs/iam/concepts/#accounts).

## Example usage

```terraform
//
// Get information about existing IAM Service Account (SA).
//
data "yandex_iam_service_account" "builder" {
  service_account_id = "aje5a**********qspd3"
}

data "yandex_iam_service_account" "deployer" {
  name = "sa_name"
}
```

## Argument reference

* `service_account_id` - (Optional) ID of a specific service account.

* `name` - (Optional) Name of a specific service account.

~> One of `service_account_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

* `description` - Description of the service account.
