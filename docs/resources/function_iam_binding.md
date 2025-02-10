---
subcategory: "Serverless Cloud Functions"
page_title: "Yandex: yandex_function_iam_binding"
description: |-
  Allows management of a single IAM binding for a Yandex Cloud Function.
---

# yandex_function_iam_binding (Resource)

## Example usage

```terraform
//
// Create a new Cloud Function IAM Binding.
//
resource "yandex_function_iam_binding" "function-iam" {
  function_id = "dns9m**********tducf"
  role        = "serverless.functions.invoker"

  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `function_id` - (Required) The [Yandex Cloud Function](https://yandex.cloud/docs/functions/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/functions/security/)

* `members` - (Required) Identities that will be granted the privilege in `role`. Each entry can have one of the following values:
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

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `id` and role.

```shell
# terraform import yandex_function_iam_binding.<resource Name> <resource Id> <resource Role>
terraform import yandex_function_iam_binding.fn_viewer dns9m**********tducf functions.viewer
```
