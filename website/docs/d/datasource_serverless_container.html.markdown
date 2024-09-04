---
layout: "yandex"
page_title: "Yandex: yandex_serverless_container"
sidebar_current: "docs-yandex-datasource-serverless-container"
description: |-
  Get information about a Yandex Cloud Serverless Container.
---

# yandex\_serverless\_container

Get information about a Yandex Cloud Serverless Container. 

```hcl
data "yandex_serverless_container" "my-container" {
  container_id = "are1samplecontainer11"
}
```

This data source is used to define Yandex Cloud Container that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `container_id` (Optional) - Yandex Cloud Serverless Container id used to define container
* `name` (Optional) - Yandex Cloud Serverless Container name used to define container
* `folder_id` (Optional) - Folder ID for the Yandex Cloud Serverless Container

~> **NOTE:** Either `container_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the Yandex Cloud Serverless Container
* `labels` - A set of key/value label pairs assigned to the Yandex Cloud Serverless Container
* `memory` - Memory in megabytes of Yandex Cloud Serverless Container
* `core` - Cores of the Yandex Cloud Serverless Container
* `core_fraction` - Core fraction (**0...100**) of the Yandex Cloud Serverless Container
* `execution_timeout` - Execution timeout (duration format) of Yandex Cloud Serverless Container
* `concurrency` - Concurrency of Yandex Cloud Serverless Container
* `service_account_id` - Service account ID of Yandex Cloud Serverless Container
* `secrets` - Secrets for Yandex Cloud Serverless Container

* `storage_mounts` - (**DEPRECATED**, use `mounts.0.object_storage` instead) Storage mounts for Yandex Cloud Serverless Container
* `storage_mounts.0.mount_point_path` - Path inside the container to access the directory in which the bucket is mounted
* `storage_mounts.0.bucket` - Name of the mounting bucket
* `storage_mounts.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted
* `storage_mounts.0.read_only` - Mount the bucket in read-only mode

* `mounts` - Mounts for Yandex Cloud Serverless Container
* `mounts.0.mount_point_path` - Path inside the container to access the directory in which the target is mounted
* `mounts.0.mode` - Mount’s accessibility mode. Valid values are `ro` and `rw`
* `mounts.0.ephemeral_disk` - One of the available mount types. Disk available during the function execution time
* `mounts.0.ephemeral_disk.0.size_gb` - Size of the ephemeral disk in GB
* `mounts.0.ephemeral_disk.0.block_size_kb` - Optional block size of the ephemeral disk in KB
* `mounts.0.object_storage` - One of the available mount types. Object storage as a mount
* `mounts.0.object_storage.0.bucket` - Name of the mounting bucket
* `mounts.0.object_storage.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted
* `image.0.url` - URL of image that deployed as Yandex Cloud Serverless Container
* `image.0.work_dir` - Working directory of Yandex Cloud Serverless Container
* `image.0.digest` - Digest of image that deployed as Yandex Cloud Serverless Container
* `image.0.command` - List of commands of the Yandex Cloud Serverless Container
* `image.0.args` - List of arguments of the Yandex Cloud Serverless Container
* `image.0.environment` -  A set of key/value environment variable pairs of Yandex Cloud Serverless Container
* `url` - Invoke URL of the Yandex Cloud Serverless Container
* `created_at` - Creation timestamp of the Yandex Cloud Serverless Container
* `revision_id` - Last revision ID of the Yandex Cloud Serverless Container
* `connectivity` - Network access. If specified the revision will be attached to specified network
* `connectivity.0.network_id` - Network the revision will have access to
* `log_options` - Options for logging from Yandex Cloud Serverless Container
