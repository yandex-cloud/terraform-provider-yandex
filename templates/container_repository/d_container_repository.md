---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Container Repository.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Container Repository. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/repository).

## Example usage

{{ tffile "examples/container_repository/d_container_repository_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - A name of the repository. The name of the repository should start with id of a container registry and match the name of the images that will be pushed in the repository. 
* `repository_id` - (Optional) The ID of a specific repository.
