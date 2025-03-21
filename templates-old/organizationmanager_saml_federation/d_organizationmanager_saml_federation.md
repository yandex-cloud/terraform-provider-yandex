---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Cloud SAML Federation.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex SAML Federation. For more information, see [the official documentation](https://yandex.cloud/docs/organization/add-federation).

## Example usage

{{ tffile "examples/organizationmanager_saml_federation/d_organizationmanager_saml_federation_1.tf" }}

## Argument Reference

The following arguments are supported:

* `federation_id` - (Optional) ID of a SAML Federation.

* `name` - (Optional) Name of a SAML Federation.

~> One of `federation_id` or `name` should be specified.

* `organization_id` - (Optional) Organization that the federation belongs to. If value is omitted, the default provider organization is used.

## Attributes Reference

* `description` - The description of the SAML Federation.
* `labels` - A set of key/value label pairs assigned to the SAML Federation.
* `issuer` - The ID of the IdP server used for authentication.
* `sso_binding` - Single sign-on endpoint binding type.
* `sso_url` - Single sign-on endpoint URL.
* `cookie_max_age` - The lifetime of a Browser cookie in seconds.
* `auto_create_account_on_login` - Indicates whether new users get added automatically on successful authentication.
* `case_insensitive_name_ids` - Indicates whether case-insensitive name ids are in use.
* `security_settings` - Federation security settings, structure is documented below.
* `created_at` - The SAML Federation creation timestamp.

---

The `security_settings` block supports:

* `encrypted_assertions` - Indicates whether encrypted assertions are enabled.
