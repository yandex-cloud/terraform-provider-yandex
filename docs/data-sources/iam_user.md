---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_user"
description: |-
  Get information about a Yandex IAM user account.
---

# yandex_iam_user (Data Source)

Get information about a Yandex IAM user account. For more information about accounts, see [Yandex Cloud IAM accounts](https://yandex.cloud/docs/iam/concepts/#accounts).

## Example usage

```terraform
//
// Get information about existing IAM User.
//
data "yandex_iam_user" "admin" {
  login = "my-yandex-login"
}
```

This data source is used to define [IAM User](https://yandex.cloud/docs/iam/concepts/#passport) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `login` (Optional) - Login name used to sign in to Yandex Passport.

* `user_id` (Optional) - User ID used to manage IAM access bindings.

~> Either `login` or `user_id` must be specified.

## Attributes Reference

The following attributes are exported:

* `user_id` - ID of IAM user account.
* `login` - Login name of IAM user account.
* `default_email` - Email address of user account.
