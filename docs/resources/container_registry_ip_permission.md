---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_registry_ip_permission"
description: |-
  Creates a new Container Registry IP Permission.
---

# yandex_container_registry_ip_permission (Resource)

Creates a new Container Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/operations/registry/registry-access)

## Example usage

```terraform
//
// Create a new Container Registry and new IP Permissions for it.
//
resource "yandex_container_registry" "my_registry" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_container_registry_ip_permission" "my_ip_permission" {
  registry_id = yandex_container_registry.my_registry.id
  push        = ["10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"]
  pull        = ["10.1.0.0/16", "10.5.0/16"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `registry_id` (String) The ID of the registry that ip restrictions applied to.

### Optional

- `pull` (Set of String) List of configured CIDRs, from which `pull` is allowed.
- `push` (Set of String) List of configured CIDRs, from which `push` is allowed.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `default` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_container_registry_ip_permission.<resource Name> <registry_id>
terraform import yandex_container_registry_ip_permission.my_ip_permission crps9**********k9psn
```
