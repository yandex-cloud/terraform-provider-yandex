---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_service_account_iam_member"
description: |-
  Allows management of a single member for a single IAM binding for a Yandex IAM service account.
---

# yandex_iam_service_account_iam_member (Resource)

**IAM policy for a service account**

When managing IAM roles, you can treat a service account either as a resource or as an identity. This resource is used to add IAM policy bindings to a service account resource to configure permissions that define who can edit the service account.

There are three different resources that help you manage your IAM policy for a service account. Each of these resources is used for a different use case:

* [yandex_iam_service_account_iam_policy](iam_service_account_iam_policy.html): Authoritative. Sets the IAM policy for the service account and replaces any existing policy already attached.
* [yandex_iam_service_account_iam_binding](iam_service_account_iam_binding.html): Authoritative for a given role. Updates the IAM policy to grant a role to a list of members. Other roles within the IAM policy for the service account are preserved.
* [yandex_iam_service_account_iam_member](iam_service_account_iam_member.html): Non-authoritative. Updates the IAM policy to grant a role to a new member. Other members for the role of the service account are preserved.

~> `yandex_iam_service_account_iam_policy` **cannot** be used in conjunction with `yandex_iam_service_account_iam_binding` and `yandex_iam_service_account_iam_member` or they will conflict over what your policy should be.

~> `yandex_iam_service_account_iam_binding` resources **can be** used in conjunction with `yandex_iam_service_account_iam_member` resources **only if** they do not grant privileges to the same role.

## Example usage

```terraform
//
// Create a new IAM Service Account IAM Member.
//
resource "yandex_iam_service_account_iam_member" "admin-account-iam" {
  service_account_id = "your-service-account-id"
  role               = "admin"
  member             = "userAccount:bar_user_id"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) The service account ID to apply a policy to.

* `role` - (Required) The role that should be applied. Only one `yandex_iam_service_account_iam_binding` can be used per role.

* `member` - (Required) Identity that will be granted the privilege in `role`. Entry can have one of the following values:
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

Service Account IAM member resource can be imported using the service account ID, resource role and member ID.

```shell
# terraform import yandex_iam_service_account_iam_member.<resource Name> "service_account_id roles/<resource Role> <member Id>"
terraform import yandex_iam_service_account_iam_member.admin-account-iam "aje5a**********qspd3 roles/admin foo@example.com"
```
