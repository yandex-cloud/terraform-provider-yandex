---
subcategory: "Organization Manager"
page_title: "Yandex: yandex_organizationmanager_saml_federation_user_account"
description: |-
  Allows management of a single SAML Federation user account within an existing Yandex.Cloud Organization.
---


# yandex_organizationmanager_saml_federation_user_account




Allows management of a single SAML Federation user account within an existing Yandex.Cloud Organization.. For more information, see [the official documentation](https://cloud.yandex.com/docs/organization/operations/federations/integration-common).

~> **Note:** If terraform user has sufficient access and user specified in data source does not exist, it will be created. This behaviour will be **deprecated** in future releases. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.

```terraform
resource "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  organization_id = "some_organization_id"
  subject_id      = "some_subject_id"
  data            = "ssh_key_data"
}
```

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
