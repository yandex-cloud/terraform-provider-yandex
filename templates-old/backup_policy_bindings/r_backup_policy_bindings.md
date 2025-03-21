---
subcategory: "Cloud Backup"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows to bind compute instance with backup policy.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud Attach and Detach VM](https://yandex.cloud/docs/backup/operations/policy-vm/attach-and-detach-vm).

~> Cloud Backup Provider must be activated in order to manipulate with policies. Active it either by UI Console or by `yc` command.

## Example usage

{{ tffile "examples/backup_policy_bindings/r_backup_policy_bindings_1.tf" }}

## Argument Reference

The following arguments are supported:

- `instance_id` (**Required**) — Compute Cloud instance ID.
- `policy_id` (**Required**) — Backup Policy ID.

## Attributes Reference

The following attributes are exported:

* `created_at` (Computed) - Creation timestamp of the Yandex Cloud Policy Bindings.
* `processing` (Computed) - Boolean flag that specifies whether the policy is in the process of binding to an instance.
* `enabled` (Computed) - Boolean flag that specifies whether the policy application is enabled. May be `false` if `processing` flag is `true`.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/backup_policy_bindings/import.sh" }}
