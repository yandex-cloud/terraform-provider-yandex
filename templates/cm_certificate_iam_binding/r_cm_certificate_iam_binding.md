---
subcategory: "Certificate Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a [Certificate](https://yandex.cloud/docs/certificate-manager/).
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single binding within IAM policy for an existing Certificate.

~> Roles controlled by `yandex_cm_certificate_iam_binding` should not be assigned using `yandex_cm_certificate_iam_member`.

~> When you delete `yandex_cm_certificate_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!

## Example usage

{{ tffile "examples/cm_certificate_iam_binding/r_cm_certificate_iam_binding_1.tf" }}

## Argument Reference

The following arguments are supported:

* `certificate_id` - (Required) The [Certificate](https://yandex.cloud/docs/certificate-manager/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/certificate-manager/security/).

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

{{ codefile "bash" "examples/cm_certificate_iam_binding/import.sh" }}
