---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_saml_federation_user_account"
description: |-
  Get information about a user of a Yandex SAML Federation.
---

# yandex_organizationmanager_saml_federation_user_account (Data Source)

Get information about a user of Yandex SAML Federation. For more information, see [the official documentation](https://yandex.cloud/docs/organization/operations/federations/integration-common).

~> If terraform user had sufficient access and user specified in data source did not exist, it would be created. This behavior will was **fixed**. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.

## Example usage

```terraform
//
// Get information about existing OrganizationManager SAML Federation User Account.
//
data "yandex_organizationmanager_saml_federation_user_account" "account" {
  federation_id = "some_federation_id"
  name_id       = "example@example.org"
}

output "my_federation.id" {
  value = data.yandex_organizationmanager_saml_federation_user_account.account.id
}
```

## Argument Reference

The following arguments are supported:

* `federation_id` - (Required) ID of a SAML Federation.

* `name_id` - (Required) Name Id of the SAML federated user.
