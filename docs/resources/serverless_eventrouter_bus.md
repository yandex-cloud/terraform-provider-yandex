---
subcategory: "Serverless Event Router"
page_title: "Yandex: yandex_serverless_eventrouter_bus"
description: |-
  Allows management of a Yandex Cloud Serverless Event Router Bus.
---

# yandex_serverless_eventrouter_bus (Resource)

Allows management of a Yandex Cloud Serverless Event Router Bus.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the bus

### Optional

- `deletion_protection` (Boolean) Deletion protection
- `description` (String) Description of the bus
- `folder_id` (String) ID of the folder that the bus belongs to
- `labels` (Map of String) Bus labels
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `cloud_id` (String) ID of the cloud that the bus resides in
- `created_at` (String) Creation timestamp
- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).