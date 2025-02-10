---
subcategory: "Managed Service for YDB"
page_title: "Yandex: yandex_ydb_database_iam_binding"
description: |-
  Allows management of a single IAM binding for a Managed service for YDB.
---

# yandex_ydb_database_iam_binding (Resource)

Allows creation and management of a single binding within IAM policy for an existing Managed YDB Database instance.

## Example usage

```terraform
//
// Create a new YDB Serverless Database and new IAM Binding for it.
//
resource "yandex_ydb_database_serverless" "database1" {
  name      = "test-ydb-serverless"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id
}

resource "yandex_ydb_database_iam_binding" "viewer" {
  database_id = yandex_ydb_database_serverless.database1.id
  role        = "ydb.viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `database_id` - (Required) The [Managed Service YDB instance](https://yandex.cloud/docs/ydb/) Database ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/ydb/security/).

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

```shell
# terraform import yandex_ydb_database_iam_binding.<resource Name> "<resource Id> <resource Role>"
terraform import yandex_lockbox_secret_iam_binding.viewer "... viewer"
```
