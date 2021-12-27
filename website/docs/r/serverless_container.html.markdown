---
layout: "yandex"
page_title: "Yandex: yandex_serverless_container"
sidebar_current: "docs-serverless-container"
description: |-
 Allows management of a Yandex Cloud Serverless Container.
---

# yandex\_serverless\_container

Allows management of Yandex Cloud Serverless Containers

## Example Usage

```hcl
resource "yandex_serverless_container" "test-container" {
  name               = "some_name"
  description        = "any description"
  memory             = 256
  execution_timeout  = "15s"
  cores              = 1
  core_fraction      = 100
  service_account_id = "are1service2account3id"
  image {
    url = "cr.yandex/yc/test-image:v1"
  }
}
```
```hcl
resource "yandex_serverless_container" "test-container-with-digest" {
 name   = "some_name"
 memory = 128
 image {
  url    = "cr.yandex/yc/test-image:v1"
  digest = "sha256:e1d772fa8795adac847a2420c87d0d2e3d38fb02f168cab8c0b5fe2fb95c47f4"
 }
}
```

## Argument Reference

The following arguments are supported:

* `name` (Required) - Yandex Cloud Serverless Container name
* `folder_id` - Folder ID for the Yandex Cloud Serverless Container
* `description` - Description of the Yandex Cloud Serverless Container
* `labels` - A set of key/value label pairs to assign to the Yandex Cloud Serverless Container

* `memory`(Required) - Memory in megabytes (**aligned to 128MB**) for Yandex Cloud Serverless Container
* `core` - Cores (**1+**) of the Yandex Cloud Serverless Container
* `core_fraction` - Core fraction (**0...100**) of the Yandex Cloud Serverless Container
* `execution_timeout` - Execution timeout in seconds (**duration format**) for Yandex Cloud Serverless Container
* `concurrency` - Concurrency of Yandex Cloud Serverless Container
* `service_account_id` - Service account ID for Yandex Cloud Serverless Container

* `image` - Revision deployment image for Yandex Cloud Serverless Container
* `image.0.url` (Required) - URL of image that will be deployed as Yandex Cloud Serverless Container
* `image.0.work_dir` - Working directory for Yandex Cloud Serverless Container
* `image.0.digest` - Digest of image that will be deployed as Yandex Cloud Serverless Container. 
  If presented, should be equal to digest that will be resolved at server side by URL. 
  Container will be updated on digest change even if `image.0.url` stays the same. 
  If field not specified then its value will be computed.
* `image.0.command` - List of commands for Yandex Cloud Serverless Container
* `image.0.args` - List of arguments for Yandex Cloud Serverless Container
* `image.0.environment` -  A set of key/value environment variable pairs for Yandex Cloud Serverless Container

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `url` - Invoke URL for the Yandex Cloud Serverless Container
* `created_at` - Creation timestamp of the Yandex Cloud Serverless Container
* `revision_id` - Last revision ID of the Yandex Cloud Serverless Container

