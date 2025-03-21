---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud IAM service account.
---

# {{.Name}} ({{.Type}})

Allows management of a Yandex Cloud IAM [service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts). To assign roles and permissions, use the [yandex_iam_service_account_iam_binding](iam_service_account_iam_binding.html), [yandex_iam_service_account_iam_member](iam_service_account_iam_member.html) and [yandex_iam_service_account_iam_policy](iam_service_account_iam_policy.html) resources.

## Example usage

{{ tffile "examples/iam_service_account/r_iam_service_account_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the service account. Can be updated without creating a new resource.

* `description` - (Optional) Description of the service account.

* `folder_id` - (Optional) ID of the folder that the service account will be created in. Defaults to the provider folder configuration.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/iam_service_account/import.sh" }}
