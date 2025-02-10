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
//
// Create a new CM Certificate IAM Member.
//
resource "yandex_cm_certificate" "your-certificate" {
  name    = "certificate-name"
  domains = ["example.com"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}

resource "yandex_cm_certificate_iam_member" "viewer_member" {
  certificate_id = yandex_cm_certificate.your-certificate.id
  role           = "viewer"

  member = "userAccount:foo_user_id"
}
```

## Argument Reference

The following arguments are supported:

* `certificate_id` - (Required) The [Certificate](https://yandex.cloud/docs/certificate-manager/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/certificate-manager/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_cm_certificate_iam_member.<resource Name> "<resource Id> <role Id> <subject id>"
terraform import yandex_cm_certificate_iam_member.viewer_member "... viewer foo@example.com"
```
