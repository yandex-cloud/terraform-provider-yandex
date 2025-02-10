---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single member for a single IAM binding for a Yandex Resource Manager folder.
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Resource Manager folder.

~> This resource *must not* be used in conjunction with `yandex_resourcemanager_folder_iam_policy` or they will conflict over what your policy should be. Similarly, roles controlled by `yandex_resourcemanager_folder_iam_binding` should not be assigned using `yandex_resourcemanager_folder_iam_member`.

## Example usage

{{ tffile "examples/resourcemanager_folder_iam_member/r_resourcemanager_folder_iam_member_1.tf" }}

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) ID of the folder to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
  * **group:{group_id}**: A unique group ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/resourcemanager_folder_iam_member/import.sh" }}
