---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a Yandex Container Repository.
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single binding within IAM policy for an existing Yandex Container Repository. For more information, see [the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/repository).

## Example usage

{{ tffile "examples/container_repository_iam_binding/r_container_repository_iam_binding_1.tf" }}

## Argument Reference

The following arguments are supported:

* `repository_id` - (Required) The [Yandex Container Repository](https://cloud.yandex.com/docs/container-registry/concepts/repository) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/container-registry/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
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

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `repository_id` and role, e.g.

```
$ terraform import yandex_container_repository_iam_binding.puller "repository_id container-registry.images.puller"
```
