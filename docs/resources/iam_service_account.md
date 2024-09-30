---
subcategory: "IAM (Identity and Access Management)"
page_title: "Yandex: yandex_iam_service_account"
description: |-
  Allows management of a Yandex.Cloud IAM service account.
---


# yandex_iam_service_account




Allows management of a Yandex.Cloud IAM [service account](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts). To assign roles and permissions, use the [yandex_iam_service_account_iam_binding](iam_service_account_iam_binding.html), [yandex_iam_service_account_iam_member](iam_service_account_iam_member.html) and [yandex_iam_service_account_iam_policy](iam_service_account_iam_policy.html) resources.

```terraform
resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = "some_sa_id"
  description        = "static access key for object storage"
  pgp_key            = "keybase:keybaseusername"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the service account. Can be updated without creating a new resource.

* `description` - (Optional) Description of the service account.

* `folder_id` - (Optional) ID of the folder that the service account will be created in. Defaults to the provider folder configuration.

## Import

A service account can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_iam_service_account.sa account_id
```
