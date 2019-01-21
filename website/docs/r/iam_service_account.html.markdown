---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account"
sidebar_current: "docs-yandex-service-account-x"
description: |-
 Allows management of a Yandex Cloud IAM service account.
---

# yandex\_service\_account

Allows management of a Yandex Cloud IAM [service account](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts).
To assign roles and permissions, use the [yandex_iam_service_account_iam_* resources](iam_service_account_iam.html).

## Example Usage

This snippet creates a service account.

```hcl
resource "yandex_iam_service_account" "sa" {
  name        = "VM Manager"
  description = "service account to manage VMs"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the service account.
    Can be updated without creating a new resource.

* `description` - (Optional) Description of the service account.

* `folder_id` - (Optional) ID of the folder that the service account will be created in.
    Defaults to the provider folder configuration.

## Import

Service accounts can be imported using their IDs, e.g.

```
$ terraform import yandex_iam_service_account.my_sa service_account_id
```
