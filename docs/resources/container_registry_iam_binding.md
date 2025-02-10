---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_registry_iam_binding"
description: |-
  Allows management of a single IAM binding for a Yandex Container Registry.
---

# yandex_container_registry_iam_binding (Resource)

Allows creation and management of a single binding within IAM policy for an existing Yandex Container Registry.

## Example usage

```terraform
//
// Create a new Container Registry and new IAM Binding for it.
//
resource "yandex_container_registry" "your-registry" {
  folder_id = "your-folder-id"
  name      = "registry-name"
}

resource "yandex_container_registry_iam_binding" "puller" {
  registry_id = yandex_container_registry.your-registry.id
  role        = "container-registry.images.puller"

  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Required) The [Yandex Container Registry](https://yandex.cloud/docs/container-registry/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/container-registry/security/).

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

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `registry_id` and role.

```bash
# terraform import yandex_container_registry_iam_binding.<binding Name> "<registry_id> <resource Role>"
terraform import yandex_container_registry_iam_binding.puller "crps9**********k9psn container-registry.images.puller"
```
