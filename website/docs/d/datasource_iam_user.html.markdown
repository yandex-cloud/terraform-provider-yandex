---
layout: "yandex"
page_title: "Yandex: yandex_iam_user"
sidebar_current: "docs-yandex-datasource-iam-user"
description: |-
  Get information about a Yandex IAM user account.
---

# yandex\_iam\_user

Get information about a Yandex IAM user account. See details about accounts [Yandex Cloud IAM users](https://cloud.yandex.com/docs/iam/concepts/users/users)

```hcl
data "yandex_iam_user" "admin" {
  login = "my-yandex-login"
}
```

This data source is used to define [IAM Users] to use to other resources.

## Argument Reference

The following arguments are supported:

* `login` (Optional) - A login name used to sign in to Yandex Passport.

* `user_id` (Optional) - User id used to manage IAM access bindings.

~> **NOTE:** One of `login` or `user_id` must be specified.

## Attributes Reference

The following attribute is exported:

* `user_id` - Id of IAM user account.
* `login` - Login name of IAM user account.
* `default_email` - Email address of user account.

[IAM Users]: https://cloud.yandex.com/docs/iam/concepts/users/users#passport
