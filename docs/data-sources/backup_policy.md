---
subcategory: "Cloud Backup"
page_title: "Yandex: yandex_backup_policy"
description: |-
  Get information about Yandex Cloud Backup Policy.
---

# yandex_backup_policy (Data Source)

Get information about a Yandex Backup Policy. For more information, see [the official documentation](https://yandex.cloud/docs/backup/concepts/policy).

## Example usage

```terraform
//
// Get information about existing Cloud Backup Policy
//
data "yandex_backup_policy" "my_policy" {
  name = "some_policy_name"
}

output "my_policy_name" {
  value = data.yandex_backup_policy.my_policy.name
}
```

## Argument Reference

The following arguments are supported:

* `policy_id` - (Optional) ID of the policy.

* `name` - (Optional) Name of the policy.

~> One of `policy_id` or `name` should be specified.

~> In case you use `name`, an error will occur if two policies with the same name exist. In this case, rename the policy or use the `policy_id`.
