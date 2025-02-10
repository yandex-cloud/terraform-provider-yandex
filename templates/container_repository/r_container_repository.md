---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new Container Repository.
---

# {{.Name}} ({{.Type}})

Creates a new container repository. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/repository).

## Example usage

{{ tffile "examples/container_repository/r_container_repository_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - A name of the repository. The name of the repository should start with id of a container registry and match the name of the images that will be pushed in the repository. 

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/container_repository/import.sh" }}
