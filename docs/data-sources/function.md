---
subcategory: "Serverless Function"
page_title: "Yandex: yandex_function"
description: |-
  Get information about a Yandex Cloud Function.
---


# yandex_function




Get information about a Yandex Cloud Function. For more information about Yandex Cloud Functions, see [Yandex Cloud Functions](https://cloud.yandex.com/docs/functions/).

```terraform
data "yandex_function_trigger" "my_trigger" {
  trigger_id = "are1sampletrigger11"
}
```

This data source is used to define [Yandex Cloud Function](https://cloud.yandex.com/docs/functions/concepts/function) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `function_id` (Optional) - Yandex Cloud Function id used to define function

* `name` (Optional) - Yandex Cloud Function name used to define function

* `folder_id` (Optional) - Folder ID for the Yandex Cloud Function

~> **NOTE:** Either `function_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the Yandex Cloud Function
* `labels` - A set of key/value label pairs to assign to the Yandex Cloud Function
* `created_at` - Creation timestamp of the Yandex Cloud Function
* `runtime` - Runtime for Yandex Cloud Function
* `entrypoint` - Entrypoint for Yandex Cloud Function
* `memory` - Memory in megabytes (**aligned to 128MB**) for Yandex Cloud Function
* `execution_timeout` - Execution timeout in seconds for Yandex Cloud Function
* `service_account_id` - Service account ID for Yandex Cloud Function
* `environment` - A set of key/value environment variables for Yandex Cloud Function
* `tags` - Tags for Yandex Cloud Function. Tag "$latest" isn't returned
* `secrets` - Secrets for Yandex Cloud Function.

* `storage_mounts` - (**DEPRECATED**, use `mounts.0.object_storage` instead) Storage mounts for Yandex Cloud Function
* `storage_mounts.0.mount_point_name` - (Required) Name of the mount point. The directory where the bucket is mounted will be accessible at the `/function/storage/<mount_point>` path
* `storage_mounts.0.bucket` - Name of the mounting bucket
* `storage_mounts.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted
* `storage_mounts.0.read_only` - Mount the bucket in read-only mode

* `mounts` - Mounts for Yandex Cloud Function.
* `mounts.0.name` - Name of the mount point. The directory where the target is mounted will be accessible at the `/function/storage/<mounts.0.name>` path
* `mounts.0.mode` - Mount’s accessibility mode. Valid values are `ro` and `rw`
* `mounts.0.ephemeral_disk` - One of the available mount types. Disk available during the function execution time
* `mounts.0.ephemeral_disk.0.size_gb` - Size of the ephemeral disk in GB
* `mounts.0.ephemeral_disk.0.block_size_kb` - Optional block size of the ephemeral disk in KB
* `mounts.0.object_storage` - One of the available mount types. Object storage as a mount
* `mounts.0.object_storage.0.bucket` - Name of the mounting bucket
* `mounts.0.object_storage.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted

* `version` - Version for Yandex Cloud Function
* `image_size` - Image size for Yandex Cloud Function
* `connectivity` - Function version connectivity. If specified the version will be attached to specified network
* `connectivity.0.network_id` - Network the version will have access to. It's essential to specify network with subnets in all availability zones
* `async_invocation` - Config for asynchronous invocations of Yandex Cloud Function
* `log_options` - Options for logging from Yandex Cloud Function
* `tmpfs_size` - Tmpfs size for Yandex Cloud Function
* `concurrency` - The maximum number of requests processed by a function instance at the same time
