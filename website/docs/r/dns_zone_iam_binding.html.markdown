---
layout: "yandex"
page_title: "Yandex: yandex_dns_zone_iam_binding"
sidebar_current: "docs-yandex-dns-zone-iam-binding"
description: |-
  Allows management of a single IAM binding for a [DNS Zone](https://cloud.yandex.com/docs/dns/).
---

## yandex\_dns\_zone\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing DNS Zone.

## Example Usage

```hcl
resource "yandex_dns_zone" "zone1" {
  name = "my-private-zone"
  zone = "example.com."
}

resource "yandex_dns_zone_iam_binding" "viewer" {
  dns_zone_id = yandex_dns_zone.zone1.id
  role        = "dns.viewer"
  members     = ["userAccount:foo_user_id"]
}
```

## Argument Reference

The following arguments are supported:

* `dns_zone_id` - (Required) The [DNS](https://cloud.yandex.com/docs/dns/) Zone ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/dns/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
    * **group:{group_id}**: A unique group ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `dns_zone_id` and role, e.g.

```
$ terraform import yandex_dns_zone_iam_binding.viewer "dns_zone_id dns.viewer"
```
