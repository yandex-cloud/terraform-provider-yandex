---
subcategory: "Serverless Containers"
page_title: "Yandex: yandex_serverless_container_iam_binding"
description: |-
  Allows management of a single IAM binding for a Yandex Serverless Container.
---

# yandex_serverless_container_iam_binding (Resource)



Allows management of a single IAM binding for a [Yandex Serverless Container](https://yandex.cloud/docs/serverless-containers/).

## Example usage

```terraform
//
// Create a new Serverless Container IAM Binding.
//
resource "yandex_serverless_container_iam_binding" "container-iam" {
  container_id = "your-container-id"
  role         = "serverless.containers.invoker"

  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `container_id` - (Required) The [Yandex Serverless Container](https://yandex.cloud/docs/serverless-containers/) ID to apply a binding to.

* `role` - (Required) The role that should be applied.

* `members` - (Required) Identities that will be granted the privilege in `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_serverless_container_iam_binding.<resource Name> "<resource Id> <resource Role>"
terraform import yandex_serverless_container_iam_binding.container-iam "... serverless.containers.invoker"
```
