---
layout: "yandex"
page_title: "Yandex: yandex_container_repository"
sidebar_current: "docs-yandex-datasource-container-repository"
description: |-
  Get information about a Yandex Container Repository.
---

# yandex\_container\_repository

Get information about a Yandex Container Repository. For more information, see
[the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/repository)

## Example Usage

```hcl
data "yandex_container_repository" "repo-1" {
  name = "some_repository_name"
}

data "yandex_container_repository" "repo-2" {
  repository_id = "some_repository_id"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the repository. The name of the repository should start with id of a container registry and match the name of the images in the repository.
* `repository_id` - (Optional) The ID of a specific repository.
