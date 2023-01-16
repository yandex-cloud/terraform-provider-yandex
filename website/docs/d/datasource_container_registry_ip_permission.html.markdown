---
layout: "yandex"
page_title: "Yandex: yandex_container_registry_ip_permission"
sidebar_current: "docs-yandex-datasource-container-registry-ip-permission"
description: |-
  Get information about a Yandex Container Registry IP Permission.
---

# yandex\_container\registry\_ip\_permission

Get information about a Yandex Container Registry IP Permission. For more information, see
[the official documentation](https://cloud.yandex.ru/docs/container-registry/operations/registry/registry-access)

## Example Usage

```hcl
resource "yandex_container_registry" "default" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
}

data "yandex_container_registry_ip_permission" "my_ip_permission_by_id" {
  registry_id = yandex_container_registry.default.id
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Optional) The ID of a specific Container Registry.

* `registry_name` - (Optional) The Name of specific Container Registry.

~> **NOTE:** Either `registry_id` or `registry_name` must be specified.

## Attributes Reference

* `push` - List of configured CIDRs, from which push is allowed.

* `pull` - List of configured CIDRs, from which pull is allowed.
