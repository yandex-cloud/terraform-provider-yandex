---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_disk_placement_group"
description: |-
  Manages a Disk Placement Group resource.
---

# yandex_compute_disk_placement_group (Resource)

A Disk Placement Group resource. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/disk#nr-disks).

## Example usage

```terraform
//
// Create a new Disk Placement Group
//
resource "yandex_compute_disk_placement_group" "group1" {
  name        = "test-pg"
  folder_id   = "abc*********123"
  description = "my description"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `name` (String) The resource name.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `zone` (String) The [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `id` (String) The ID of this resource.
- `status` (String) Status of the Disk Placement Group.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_compute_disk_placement_group.<resource Name> <resource Id>
terraform import yandex_compute_disk_placement_group.my_disk_group ...
```
