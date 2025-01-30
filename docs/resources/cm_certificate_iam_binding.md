---
subcategory: "Certificate Manager"
page_title: "Yandex: yandex_cm_certificate_iam_binding"
description: |-
  Allows management of a single IAM binding for a [Certificate](https://yandex.cloud/docs/certificate-manager/).
---

# yandex_cm_certificate_iam_binding (Resource)

Allows creation and management of a single binding within IAM policy for an existing Certificate.

~> Roles controlled by `yandex_cm_certificate_iam_binding` should not be assigned using `yandex_cm_certificate_iam_member`.

~> When you delete `yandex_cm_certificate_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!

## Example usage

```terraform
resource "yandex_cm_certificate" "your-certificate" {
  name = "certificate-name"
  domains = ["example.com"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}

resource "yandex_cm_certificate_iam_binding" "viewer" {
  certificate_id = yandex_cm_certificate.your-certificate.id
  role      = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `certificate_id` - (Required) The [Certificate](https://yandex.cloud/docs/certificate-manager/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/certificate-manager/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `certificate_id` and role, e.g.

```
$ terraform import yandex_cm_certificate_iam_binding.viewer "certificate_id viewer"
```
