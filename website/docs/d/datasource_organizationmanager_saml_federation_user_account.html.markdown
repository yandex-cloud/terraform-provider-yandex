---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_saml_federation_user_account"
sidebar_current: "docs-yandex-datasource-organizationmanager-saml-federation-user-account"
description: |-
  Get information about a user of a Yandex SAML Federation.
---

# yandex\_organizationmanager\_saml\_federation\_user\_account

Get information about a user of Yandex SAML Federation. For more information, see
[the official documentation](https://cloud.yandex.com/docs/organization/operations/federations/integration-common).

## Example Usage

```hcl
data "yandex_organizationmanager_saml_federation_user_account" account {
  federation_id = "some_federation_id"
  name_id       = "example@example.org"
}

output "my_federation.id" {
  value = "${data.yandex_organizationmanager_saml_federation_user_account.account.id}"
}
```

## Argument Reference

The following arguments are supported:

* `federation_id` - (Required) ID of a SAML Federation.

* `name_id` - (Required) Name Id of the SAML federated user.
