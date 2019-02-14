---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account_key"
sidebar_current: "docs-yandex-iam-service-account-key"
description: |-
 Allows management of a Yandex Cloud IAM service account static key.
---

# yandex\_iam\_service\_account\_key

Allows management of a [Yandex Cloud IAM service account static key](https://cloud.yandex.com/docs/storage/operations/security/get-static-key).
This key is used to access [Yandex Object Storage].

## Example Usage

This snippet creates a service account static key.

```hcl
resource "yandex_iam_service_account_key" "sa-key" {
  service_account_id = "some_sa_id"
  description        = "key to access primary storage"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of the service account which is used to get a static key.

- - -

* `description` - (Optional) The description of the service account.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `access_key` - The access key used to access [Yandex Object Storage].

* `secret_key` - The secret key used to access [Yandex Object Storage].

[Yandex Object Storage]: https://cloud.yandex.com/docs/storage/