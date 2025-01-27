---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new Container Repository.
---

# {{.Name}} ({{.Type}})

Creates a new container repository. For more information, see [the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/repository).

## Example usage

{{ tffile "examples/container_repository/r_container_repository_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - A name of the repository. The name of the repository should start with id of a container registry and match the name of the images that will be pushed in the repository. 

## Import

A repository can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_container_repository.my-repository repository_id
```
