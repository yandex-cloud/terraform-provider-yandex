---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account"
sidebar_current: "docs-yandex-iam-service-account-x"
description: |-
 Allows management of a Yandex.Cloud IAM service account.
---

# yandex\_iam\_service\_account

Allows management of a Yandex.Cloud IAM [service account](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts).
To assign roles and permissions, use the [yandex_iam_service_account_iam_binding](iam_service_account_iam_binding.html), 
[yandex_iam_service_account_iam_member](iam_service_account_iam_member.html) and 
[yandex_iam_service_account_iam_policy](iam_service_account_iam_policy.html) resources.

## Example Usage

This snippet creates a service account.

```hcl
resource "yandex_iam_service_account" "sa" {
  name        = "vmmanager"
  description = "service account to manage VMs"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the service account.
    Can be updated without creating a new resource.

* `description` - (Optional) Description of the service account.

* `folder_id` - (Optional) ID of the folder that the service account will be created in.
    Defaults to the provider folder configuration.

## Import

A service account can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_iam_service_account.sa account_id
```
