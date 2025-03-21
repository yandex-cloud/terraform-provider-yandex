---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a user of a Yandex SAML Federation.
---

# {{.Name}} ({{.Type}})

Get information about a user of Yandex SAML Federation. For more information, see [the official documentation](https://yandex.cloud/docs/organization/operations/federations/integration-common).

~> If terraform user had sufficient access and user specified in data source did not exist, it would be created. This behavior will was **fixed**. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.

## Example usage

{{ tffile "examples/organizationmanager_saml_federation_user_account/d_organizationmanager_saml_federation_user_account_1.tf" }}

## Argument Reference

The following arguments are supported:

* `federation_id` - (Required) ID of a SAML Federation.

* `name_id` - (Required) Name Id of the SAML federated user.
