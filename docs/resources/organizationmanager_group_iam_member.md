---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_group_iam_member"
description: |-
  Allows management of a single member for a single IAM binding on a Yandex Cloud Organization Manager Group.
---

# yandex_organizationmanager_group_iam_member (Resource)

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Organization Manager Group.

## Example usage

```terraform
resource "yandex_organizationmanager_group_iam_member" "editor" {
  group_id = "some_group_id"
  role     = "editor"
  member   = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required) ID of the organization to attach a policy to.

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

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account. This member resource can be imported using the `group_id`, role, and account, e.g.

```
$ terraform import yandex_organizationmanager_group_iam_member.my_project "group_id viewer foo@example.com"
```
