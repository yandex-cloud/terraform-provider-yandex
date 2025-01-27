---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single SAML Federation user account within an existing Yandex Cloud Organization.
---

# {{.Name}} ({{.Type}})

Allows management of a single SAML Federation user account within an existing Yandex Cloud Organization.. For more information, see [the official documentation](https://cloud.yandex.com/docs/organization/operations/federations/integration-common).

~> If terraform user has sufficient access and user specified in data source does not exist, it will be created. This behaviour will be **deprecated** in future releases. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.

## Example usage

{{ tffile "examples/organizationmanager_saml_federation_user_account/r_organizationmanager_saml_federation_user_account_1.tf" }}

## Argument Reference

The following arguments are supported:

* `federation_id` - (Required) ID of a SAML Federation.

* `name_id` - (Required) Name ID of the SAML federated user.
* 

## Import

A Yandex SAML Federation user account can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_organizationmanager_saml_federation_user_account.account "user_id"
```
