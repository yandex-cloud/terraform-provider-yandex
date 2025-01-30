---
subcategory: "Certificate Manager"
page_title: "Yandex: yandex_cm_certificate_iam_member"
description: |-
  Allows management of a single member for a single IAM binding for a [Certificate](https://yandex.cloud/docs/certificate-manager/).
---

# yandex_cm_certificate_iam_member (Resource)

Allows creation and management of a single member for a single binding within the IAM policy for an existing Certificate.

~> Roles controlled by `yandex_cm_certificate_iam_binding` should not be assigned using `yandex_cm_certificate_iam_member`.

## Example usage

```terraform
resource "yandex_cm_certificate" "your-certificate" {
  name = "certificate-name"
  domains = ["example.com"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}

resource "yandex_cm_certificate_iam_member" "viewer" {
  certificate_id = yandex_cm_certificate.your-certificate.id
  role      = "viewer"

  member = "userAccount:foo_user_id"
}
```

## Argument Reference

The following arguments are supported:

* `certificate_id` - (Required) The [Certificate](https://yandex.cloud/docs/certificate-manager/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/certificate-manager/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account. This member resource can be imported using the `certificate_id`, role, and account, e.g.

```
$ terraform import yandex_cm_certificate_iam_member.viewer "certificate_id viewer foo@example.com"
```
