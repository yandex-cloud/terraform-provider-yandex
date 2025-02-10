---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new Container Registry IP Permission.
---

# {{.Name}} ({{.Type}})

Creates a new Container Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/operations/registry/registry-access)

## Example usage

{{ tffile "examples/container_registry_ip_permission/r_container_registry_ip_permission_1.tf" }}

## Argument Reference

The following arguments are supported:

* `registry_id` - (Required) The ID of the registry that ip restrictions applied to.

* `push` - List of configured CIDRs, from which push is allowed.

* `pull` - List of configured CIDRs, from which pull is allowed.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/container_registry_ip_permission/import.sh" }}
