---
subcategory: "Cloud Backup"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about Yandex Cloud Backup Policy.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Backup Policy. For more information, see [the official documentation](https://yandex.cloud/docs/backup/concepts/policy).

## Example usage

{{ tffile "examples/backup_policy/d_backup_policy_1.tf" }}

## Argument Reference

The following arguments are supported:

* `policy_id` - (Optional) ID of the policy.

* `name` - (Optional) Name of the policy.

~> One of `policy_id` or `name` should be specified.

~> In case you use `name`, an error will occur if two policies with the same name exist. In this case, rename the policy or use the `policy_id`.
