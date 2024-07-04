---
layout: "yandex"
page_title: "Yandex: yandex_backup_policy"
sidebar_current: "docs-yandex-datasource-backup-policy"
description: |-
Get information about a Yandex Backup Policy.
---

# yandex\_compute\_filesystem

Get information about a Yandex Backup Policy. For more information, see
[the official documentation](https://yandex.cloud/docs/backup/concepts/policy).

## Example Usage

```hcl
data "yandex_backup_policy" "my_policy" {
  policy_id = "some_policy_id"
}

output "my_policy_name" {
  value = data.yandex_backup_policy.my_policy.name
}
```

## Argument Reference

The following arguments are supported:

* `policy_id` - (Required) ID of the policy.
