---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_service_account"
description: |-
  Allows management of a Yandex Cloud IAM service account.
---

# yandex_iam_service_account (Resource)

Allows management of a Yandex Cloud IAM [service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts). To assign roles and permissions, use the [yandex_iam_service_account_iam_binding](iam_service_account_iam_binding.html), [yandex_iam_service_account_iam_member](iam_service_account_iam_member.html) and [yandex_iam_service_account_iam_policy](iam_service_account_iam_policy.html) resources.

## Example usage

```terraform
//
// Create a new IAM Service Account (SA).
//
resource "yandex_iam_service_account" "builder" {
  name        = "vmmanager"
  description = "service account to manage VMs"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the service account. Can be updated without creating a new resource.

* `description` - (Optional) Description of the service account.

* `folder_id` - (Optional) ID of the folder that the service account will be created in. Defaults to the provider folder configuration.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_iam_service_account.<resource Name> <resource Id>
terraform import yandex_iam_service_account.builder aje5a**********qspd3
```
