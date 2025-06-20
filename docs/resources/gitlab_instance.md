---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "yandex_gitlab_instance Resource - yandex"
subcategory: ""
description: |-
  Managed Gitlab instance.
---

# yandex_gitlab_instance (Resource)

Managed Gitlab instance.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `admin_email` (String) An email of admin user in Gitlab.
- `admin_login` (String) A login of admin user in Gitlab.
- `approval_rules_id` (String) Approval rules configuration. One of: NONE, BASIC, STANDARD, ADVANCED.
- `backup_retain_period_days` (Number) Auto backups retain period in days.
- `disk_size` (Number) Amount of disk storage available to a instance in GB.
- `domain` (String) Domain of the Gitlab instance.
- `name` (String) The resource name.
- `resource_preset_id` (String) ID of the preset for computational resources available to the instance (CPU, memory etc.). One of: s2.micro, s2.small, s2.medium, s2.large.
- `subnet_id` (String) ID of the subnet where the GitLab instance is located.

### Optional

- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `maintenance_delete_untagged` (Boolean) The `true` value means that untagged images will be deleted during maintenance.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `gitlab_version` (String) Version of Gitlab on instance.
- `id` (String) The resource identifier.
- `status` (String) Status of the instance.
- `updated_at` (String) The timestamp when the instance was updated.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
