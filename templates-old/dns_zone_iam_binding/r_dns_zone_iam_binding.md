---
subcategory: "Cloud Domain Name System (DNS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a Cloud DNS Zone.
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single binding within IAM policy for an existing DNS Zone.

## Example usage

{{ tffile "examples/dns_zone_iam_binding/r_dns_zone_iam_binding_1.tf" }}

## Argument Reference

The following arguments are supported:

* `dns_zone_id` - (Required) The [DNS](https://yandex.cloud/docs/dns/) Zone ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/dns/security/).

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

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `dns_zone_id` and role.

{{ codefile "shell" "examples/dns_zone_iam_binding/import.sh" }}
