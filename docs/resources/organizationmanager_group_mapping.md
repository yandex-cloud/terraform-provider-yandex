---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_group_mapping"
description: |-
  Allows management of a Yandex Cloud Organization Manager Group Mapping.
---

# yandex_organizationmanager_group_mapping (Resource)

Allows management of [Yandex Cloud Organization Manager Group Mapping](https://yandex.cloud/docs/organization/concepts/add-federation#group-mapping). It supports the creation, updating(enabling/disabling), and deletion of group mapping.

## Example Usage

```terraform
//
// Create a new OrganizationManager Group Mapping.
//
resource "yandex_organizationmanager_group_mapping" "my_group_map" {
  federation_id = "my-federation-id"
  enabled       = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `federation_id` (String) ID of the SAML Federation.

### Optional

- `enabled` (Boolean) Set "true" to enable organization manager group mapping.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).




## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_organizationmanager_group_mapping.<resource Name> <resource Id>
terraform import yandex_organizationmanager_group.my_group_map ...
```