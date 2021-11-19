---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_saml_federation"
sidebar_current: "docs-yandex-datasource-organizationmanager-saml-federation"
description: |-
  Get information about a Yandex.Cloud SAML Federation.
---

# yandex\_organizationmanager\_saml\_federation

Get information about a Yandex SAML Federation. For more information, see
[the official documentation](https://cloud.yandex.com/docs/organization/add-federation).

## Example Usage

```hcl
data "yandex_organizationmanager_saml_federation" federation {
  federation_id   = "some_federation_id"
  organization_id = "some_organization_id"
}

output "my_federation.name" {
  value = "${data.yandex_organizationmanager_saml_federation.federation.name}"
}
```

## Argument Reference

The following arguments are supported:

* `federation_id` - (Optional) ID of a SAML Federation.

* `name` - (Optional) Name of a SAML Federation.

~> **NOTE:** One of `federation_id` or `name` should be specified.

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
