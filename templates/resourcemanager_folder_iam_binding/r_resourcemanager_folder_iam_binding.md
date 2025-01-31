---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a Yandex Resource Manager folder.
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single binding within IAM policy for an existing Yandex Resource Manager folder.

~> This resource *must not* be used in conjunction with `yandex_resourcemanager_folder_iam_policy` or they will conflict over what your policy should be.

~> When you delete `yandex_resourcemanager_folder_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!

## Example usage

{{ tffile "examples/resourcemanager_folder_iam_binding/r_resourcemanager_folder_iam_binding_1.tf" }}

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) ID of the folder to attach a policy to.

* `role` - (Required) The role that should be assigned. Only one `yandex_resourcemanager_folder_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege that is specified in the `role` field. Each entry can have one of the following values:
  * **userAccount:{user_id}**: An email address that represents a specific Yandex account. For example, ivan@yandex.ru or joe@example.com.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
  * **group:{group_id}**: A unique group ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `folder_id` and role, e.g.

```
$ terraform import yandex_resourcemanager_folder_iam_binding.viewer "folder_id viewer"
```
