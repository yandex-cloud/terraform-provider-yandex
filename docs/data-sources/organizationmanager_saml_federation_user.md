---
subcategory: "Identity Hub"
---

# yandex_organizationmanager_saml_federation_user (DataSource)

Get information about a user of Yandex SAML Federation by their IAM user account ID. For more information, see [the official documentation](https://yandex.cloud/docs/organization/operations/federations/integration-common).

## Example usage

```terraform
//
// Get information about existing OrganizationManager SAML Federation User by their IAM user account ID.
//
data "yandex_organizationmanager_saml_federation_user" "account" {
  federation_id   = "some_federation_id"
  user_account_id = "some_user_account_id"
}

output "my_federation_user.name_id" {
  value = data.yandex_organizationmanager_saml_federation_user.account.name_id
}

output "my_federation_user.preferred_username" {
  value = data.yandex_organizationmanager_saml_federation_user.account.preferred_username
}
```

## Arguments & Attributes Reference

- `federation_id` (**Required**)(String). ID of a SAML Federation.
- `user_account_id` (**Required**)(String). ID of the IAM user account.
- `attributes` (*Read-Only*) (Map Of String). Additional SAML attributes of the user, as sent by the identity provider in the SAML assertion. If an attribute has multiple values, they are joined with a comma.
- `email` (*Read-Only*) (String). Email of the user, as provided by the identity provider.
- `family_name` (*Read-Only*) (String). Family (last) name of the user, as provided by the identity provider.
- `given_name` (*Read-Only*) (String). Given (first) name of the user, as provided by the identity provider.
- `id` (String). 
- `name` (*Read-Only*) (String). Full name of the user, as provided by the identity provider.
- `name_id` (*Read-Only*) (String). Name ID of the SAML federated user.
- `preferred_username` (*Read-Only*) (String). Username of the user, as provided by the identity provider.
