---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_policy"
description: |-
  Generates an IAM policy that can be referenced by other resources and applied to them.
---


# yandex_iam_policy




Generates an [IAM](https://cloud.yandex.com/docs/iam/) policy document that may be referenced by and applied to other Yandex.Cloud Platform resources, such as the `yandex_resourcemanager_folder` resource.

## Example usage

```terraform
data "yandex_iam_policy" "admin" {
  binding {
    role = "admin"

    members = [
      "userAccount:user_id_1"
    ]
  }

  binding {
    role = "viewer"

    members = [
      "userAccount:user_id_2"
    ]
  }
}
```

This data source is used to define [IAM](https://cloud.yandex.com/docs/iam/) policies to apply to other resources. Currently, defining a policy through a data source and referencing that policy from another resource is the only way to apply an IAM policy to a resource.

## Argument Reference

The following arguments are supported:

* `binding` (Required) - A nested configuration block (described below) that defines a binding to be included in the policy document. Multiple `binding` arguments are supported.

Each policy document configuration must have one or more `binding` blocks. Each block accepts the following arguments:

* `role` (Required) - The role/permission that will be granted to the members. See the [IAM Roles](https://cloud.yandex.com/docs/iam/concepts/access-control/roles) documentation for a complete list of roles.

* `members` (Required) - An array of identities that will be granted the privilege in the `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
  * **group:{group_id}**: A unique group ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Attributes Reference

The following attribute is exported:

* `policy_data` - The above bindings serialized in a format suitable for referencing from a resource that supports IAM.
