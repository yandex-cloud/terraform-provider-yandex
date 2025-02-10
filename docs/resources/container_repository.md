---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_repository"
description: |-
  Creates a new Container Repository.
---

# yandex_container_repository (Resource)

Creates a new container repository. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/repository).

## Example usage

```terraform
//
// Create a new Container Registry and new Repository with it.
//
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

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_container_repository.<resource Name> <repository_id>
terraform import yandex_container_repository.my-repository crps9**********k9psn
```
