---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account_static_access_key"
sidebar_current: "docs-yandex-iam-service-account-static-access-key"
description: |-
 Allows management of a Yandex.Cloud IAM service account static access key.
---

# yandex\_iam\_service\_account\_static\_access\_key

Allows management of a [Yandex.Cloud IAM service account static access keys](https://cloud.yandex.com/docs/iam/operations/sa/create-access-key).
Generated pair of keys are used to access [Yandex Object Storage] on behalf of service account.

Before use keys do not forget to [assign a proper role](https://cloud.yandex.com/docs/iam/operations/sa/assign-role-for-sa) to a service account.

## Example Usage

This snippet creates a service account static access key.

```hcl
resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = "some_sa_id"
  description        = "static access key for object storage"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of the service account which is used to get a static key.

- - -

* `description` - (Optional) The description of the service account static key.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `access_key` - ID of the static access key.

* `secret_key` - Private part of generated static access key. 

* `created_at` - Creation timestamp of the static access key.

[Yandex Object Storage]: https://cloud.yandex.com/docs/storage/
