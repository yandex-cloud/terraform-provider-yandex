---
subcategory: "Organization Manager"
page_title: "Yandex: yandex_organizationmanager_saml_federation_user_account"
description: |-
  Get information about a user of a Yandex SAML Federation.
---


# yandex_organizationmanager_saml_federation_user_account




Get information about a user of Yandex SAML Federation. For more information, see [the official documentation](https://cloud.yandex.com/docs/organization/operations/federations/integration-common).

~> **Note:** If terraform user had sufficient access and user specified in data source did not exist, it would be created. This behaviour will was **fixed**. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.

```terraform
data "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  user_ssh_key_id = "some_user_ssh_key_id"
}

output "my_user_ssh_key_name" {
  value = "data.yandex_organizationmanager_user_ssh_key.my_user_ssh_key.name"
}
```

## Argument Reference

The following arguments are supported:

* `federation_id` - (Required) ID of a SAML Federation.

* `name_id` - (Required) Name Id of the SAML federated user.
