---
subcategory: "Organization Manager"
page_title: "Yandex: yandex_organizationmanager_organization_iam_binding"
description: |-
  Allows management of a single IAM binding for a Yandex Organization Manager organization.
---


# yandex_organizationmanager_organization_iam_binding




Allows creation and management of a single binding within IAM policy for an existing Yandex.Cloud Organization Manager organization.

```terraform
resource "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  organization_id = "some_organization_id"
  subject_id      = "some_subject_id"
  data            = "ssh_key_data"
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) ID of the organization to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one `yandex_organizationmanager_organization_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege in the `role`. Each entry can have one of the following values:
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

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `organization_id` and role, e.g.

```
$ terraform import yandex_organizationmanager_organization_iam_binding.viewer "organization_id viewer"
```
