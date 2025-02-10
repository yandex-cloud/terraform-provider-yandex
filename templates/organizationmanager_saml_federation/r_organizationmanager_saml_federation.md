---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single SAML Federation within an existing Yandex Cloud Organization.
---

# {{.Name}} ({{.Type}})

Allows management of a single SAML Federation within an existing Yandex Cloud Organization.

## Example usage

{{ tffile "examples/organizationmanager_saml_federation/r_organizationmanager_saml_federation_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SAML Federation.
* `description` - (Optional) The description of the SAML Federation.
* `organization_id` - (Required, Forces new resource) The organization to attach this SAML Federation to.
* `labels` - (Optional) A set of key/value label pairs assigned to the SAML Federation.
* `issuer` - (Required) The ID of the IdP server to be used for authentication. The IdP server also responds to IAM with this ID after the user authenticates.
* `sso_binding` - (Required) Single sign-on endpoint binding type. Most Identity Providers support the `POST` binding type. SAML Binding is a mapping of a SAML protocol message onto standard messaging formats and/or communications protocols.
* `sso_url` - (Required) Single sign-on endpoint URL. Specify the link to the IdP login page here.
* `cookie_max_age` - (Optional, Computed) The lifetime of a Browser cookie in seconds. If the cookie is still valid, the management console authenticates the user immediately and redirects them to the home page. The default value is `8h`.
* `auto_create_account_on_login` - (Optional, Computed) Add new users automatically on successful authentication. The user will get the `resource-manager.clouds.member` role automatically, but you need to grant other roles to them. If the value is `false`, users who aren't added to the cloud can't log in, even if they have authenticated on your server.
* `case_insensitive_name_ids` - (Optional, Computed) Use case-insensitive name ids.
* `security_settings` - (Optional, Computed) Federation security settings, structure is documented below.

---

The `security_settings` block supports:

* `encrypted_assertions` - (Optional, Computed) Enable encrypted assertions.
* `force_authn` - (Optional, Computed) - Force authentication on session expiration

## Attributes Reference

* `created_at` - (Computed) The SAML Federation creation timestamp.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/organizationmanager_saml_federation/import.sh" }}
