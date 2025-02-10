---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud_iam_member"
description: |-
  Allows management of a single member for a single IAM binding on a Yandex Resource Manager cloud.
---

# yandex_resourcemanager_cloud_iam_member (Resource)

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Resource Manager cloud.

~> Roles controlled by `yandex_resourcemanager_cloud_iam_binding` should not be assigned using `yandex_resourcemanager_cloud_iam_member`.

~> When you delete `yandex_resourcemanager_cloud_iam_binding` resource, the roles can be deleted from other users within the cloud as well. Be careful!

## Example usage

```terraform
//
// Create a new Cloud and new IAM Member for it.
//
data "yandex_resourcemanager_cloud" "department1" {
  name = "Department 1"
}

resource "yandex_resourcemanager_cloud_iam_member" "admin" {
  cloud_id = data.yandex_resourcemanager_cloud.department1.id
  role     = "editor"
  member   = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Required) ID of the cloud to attach a policy to.

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

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_resourcemanager_cloud_iam_member.<resource Name> "<resource Id> <resource Role> <subject>"
terraform import yandex_resourcemanager_cloud_iam_member.admin "b1g3a**********qc5df admin foo@example.com"
```
