---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of the IAM policy for a Yandex Resource Manager folder.
---

# {{.Name}} ({{.Type}})

Allows creation and management of the IAM policy for an existing Yandex Resource Manager folder.

## Example usage

{{ tffile "examples/resourcemanager_folder_iam_policy/r_resourcemanager_folder_iam_policy_1.tf" }}

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) ID of the folder that the policy is attached to.
* `policy_data` - (Required) The `yandex_iam_policy` data source that represents the IAM policy that will be applied to the folder. This policy overrides any existing policy applied to the folder.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/resourcemanager_folder_iam_policy/import.sh" }}
