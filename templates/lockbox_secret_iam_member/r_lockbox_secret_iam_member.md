---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single member for a single IAM binding for a Lockbox Secret.
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Lockbox Secret.

~> Roles controlled by `yandex_lockbox_secret_iam_binding` should not be assigned using `yandex_lockbox_secret_iam_member`.

## Example usage

{{ tffile "examples/lockbox_secret_iam_member/r_lockbox_secret_iam_member_1.tf" }}

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The [Yandex Lockbox Secret](https://cloud.yandex.com/docs/lockbox/) Secret ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/lockbox/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account. This member resource can be imported using the `secret_id`, role, and account, e.g.

```
$ terraform import {{.Name}}.viewer "secret_id viewer foo@example.com"
```
