---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_repository"
description: |-
  Creates a new Container Repository.
---

# yandex_container_repository (Resource)

Creates a new container repository. For more information, see [the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/repository).

## Example usage

```terraform
resource "yandex_container_registry" "my-registry" {
  name = "test-registry"
}

resource "yandex_container_repository" "my-repository" {
  name = "${yandex_container_registry.my-registry.id}/test-repository"
}
```

## Argument Reference

The following arguments are supported:

* `name` - A name of the repository. The name of the repository should start with id of a container registry and match the name of the images that will be pushed in the repository. 

## Import

A repository can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_container_repository.my-repository repository_id
```
