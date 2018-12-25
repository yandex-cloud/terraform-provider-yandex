---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account_key"
sidebar_current: "docs-yandex-service-account-key"
description: |-
 Allows management of a Yandex Cloud IAM service account static key.
---

# yandex\_service\_account\_key

Allows management of a [Yandex Cloud IAM service account static key](https://cloud.yandex.com/docs/storage/operations/security/get-static-key).
This key is used to access [Yandex Object Storage](https://cloud.yandex.com/docs/storage/).

## Example Usage

This snippet creates a service account static key.

```hcl
resource "yandex_iam_service_account_key" "sa-key" {
  name        = "primary key"
  description = "key to access S3"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of service account which is used to get a static key.

- - -

* `name` - (Optional) Name of the service account.
    Can be updated without creating a new resource.

* `description` - (Optional) The description of the service account.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `access_key` - The access key used to access S3.

* `secret_key` - The secret key used to access S3.
