---
layout: "yandex"
page_title: "Yandex: yandex_container_registry_ip_permission"
sidebar_current: "docs-yandex-container-registry-ip-permission"
description: |-
  Creates a new Container Registry IP Permission.
---

# yandex\_container\_registry\_ip\_permission

Creates a new Container Registry IP Permission. For more information, see
[the official documentation](https://cloud.yandex.ru/docs/container-registry/operations/registry/registry-access)

## Example Usage

```hcl
resource "yandex_container_registry" "my_registry" {
  name = "test-registry"
}

resource "yandex_container_registry_ip_permission" "my_ip_permission" {
  registry_id = yandex_container_registry.my_registry.id
  push        = [ "10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16" ]
  pull        = [ "10.1.0.0/16", "10.5.0/16" ]
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Required) The ID of the registry that ip restrictions applied to.

* `push` - List of configured CIDRs, from which push is allowed.

* `pull` - List of configured CIDRs, from which pull is allowed.

## Import

An ip premission can be imported using the `id` of the Container Registry it is applied to, e.g.

```bash
terraform import yandex_container_registry_ip_permission.my_ip_permission registry_id
```
