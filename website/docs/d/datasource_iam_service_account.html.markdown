---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account"
sidebar_current: "docs-yandex-datasource-iam-service-account"
description: |-
  Get information about a Yandex IAM service account.
---

# yandex\_iam\_service\_account

Get information about a Yandex IAM service account. For more information about accounts, see 
[Yandex.Cloud IAM users](https://cloud.yandex.com/docs/iam/concepts/users/users).

```hcl
data "yandex_iam_service_account" "builder" {
  service_account_id = "sa_id"
}
```

## Argument reference

* `name` - Name of the service account.
    Can be updated without creating a new resource.

* `description` - Description of the service account.

* `folder_id` - ID of the folder that the service account will be created in.
    If omitted, the provider folder configuration is used by default.
