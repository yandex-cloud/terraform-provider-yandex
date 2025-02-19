---
subcategory: "Serverless Cloud Functions"
page_title: "Yandex: yandex_function"
description: |-
  Allows management of a Yandex Cloud Function.
---

# yandex_function (Resource)

Allows management of [Yandex Cloud Function](https://yandex.cloud/docs/functions)

## Example usage

```terraform
//
// Create a new Yandex Cloud Function
//
resource "yandex_function" "test-function" {
  name               = "some_name"
  description        = "any description"
  user_hash          = "any_user_defined_string"
  runtime            = "python37"
  entrypoint         = "main"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = "ajeih**********838kk"
  tags               = ["my_tag"]
  secrets {
    id                   = yandex_lockbox_secret.secret.id
    version_id           = yandex_lockbox_secret_version.secret_version.id
    key                  = "secret-key"
    environment_variable = "ENV_VARIABLE"
  }
  content {
    zip_filename = "function.zip"
  }
  mounts {
    name = "mnt"
    ephemeral_disk {
      size_gb = 32
    }
  }
  async_invocation {
    retries_count       = "3"
    service_account_id = "ajeih**********838kk"
    ymq_failure_target {
      service_account_id = "ajeqr**********qb76m"
      arn                = "yrn:yc:ymq:ru-central1:b1glr**********9hsfp:fail"
    }
    ymq_success_target {
      service_account_id = "ajeqr**********qb76m"
      arn                = "yrn:yc:ymq:ru-central1:b1glr**********9hsfp:success"
    }
  }
  log_options {
    log_group_id = "e2392**********eq9fr"
    min_level    = "ERROR"
  }
}
```

```terraform
//
// Create a new Yandex Cloud Function with mounted Object Storage Bucket.
//
resource "yandex_function" "test-function" {
  name               = "some_name"
  user_hash          = "v1"
  runtime            = "python37"
  entrypoint         = "index.handler"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = yandex_iam_service_account.sa.id
  content {
    zip_filename = "function.zip"
  }
  mounts {
    name = "mnt"
    mode = "ro"
    object_storage {
      bucket = yandex_storage_bucket.my-bucket.bucket
    }
  }
}

locals {
  folder_id = "folder_id"
}

resource "yandex_iam_service_account" "sa" {
  folder_id = local.folder_id
  name      = "test-sa"
}

resource "yandex_resourcemanager_folder_iam_member" "sa-editor" {
  folder_id = local.folder_id
  role      = "storage.editor"
  member    = "serviceAccount:${yandex_iam_service_account.sa.id}"
}

resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = yandex_iam_service_account.sa.id
  description        = "static access key for object storage"
}

resource "yandex_storage_bucket" "my-bucket" {
  access_key = yandex_iam_service_account_static_access_key.sa-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-static-key.secret_key
  bucket     = "bucket"
}
```

## Argument Reference

The following arguments are supported:

* `name` (Required) - Yandex Cloud Function name used to define trigger
* `folder_id` - Folder ID for the Yandex Cloud Function
* `description` - Description of the Yandex Cloud Function
* `labels` - A set of key/value label pairs to assign to the Yandex Cloud Function
* `user_hash` - (Required) User-defined string for current function version. User must change this string any times when function changed. Function will be updated when hash is changed.

* `runtime` - (Required) Runtime for Yandex Cloud Function
* `entrypoint` - (Required) Entrypoint for Yandex Cloud Function
* `memory` - (Required) Memory in megabytes (**aligned to 128MB**) for Yandex Cloud Function
* `execution_timeout` - Execution timeout in seconds for Yandex Cloud Function
* `service_account_id` - Service account ID for Yandex Cloud Function
* `environment` - A set of key/value environment variables for Yandex Cloud Function. Each key must begin with a letter (A-Z, a-z).
* `tags` - Tags for Yandex Cloud Function. Tag "$latest" isn't returned
* `secrets` - Secrets for Yandex Cloud Function.

* `storage_mounts` - (**DEPRECATED**, use `mounts.0.object_storage` instead) Storage mounts for Yandex Cloud Function
* `storage_mounts.0.mount_point_name` - (Required) Name of the mount point. The directory where the bucket is mounted will be accessible at the `/function/storage/<mount_point>` path
* `storage_mounts.0.bucket` - (Required) Name of the mounting bucket
* `storage_mounts.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted
* `storage_mounts.0.read_only` - Mount the bucket in read-only mode

* `mounts` - Mounts for Yandex Cloud Function.
* `mounts.0.name` - (Required) Name of the mount point. The directory where the target is mounted will be accessible at the `/function/storage/<mounts.0.name>` path
* `mounts.0.mode` - Mount’s accessibility mode. Valid values are `ro` and `rw`
* `mounts.0.ephemeral_disk` - One of the available mount types. Disk available during the function execution time
* `mounts.0.ephemeral_disk.0.size_gb` - (Required) Size of the ephemeral disk in GB
* `mounts.0.ephemeral_disk.0.block_size_kb` - Optional block size of the ephemeral disk in KB
* `mounts.0.object_storage` - One of the available mount types. Object storage as a mount
* `mounts.0.object_storage.0.bucket` - (Required) Name of the mounting bucket
* `mounts.0.object_storage.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted

* `version` - Version for Yandex Cloud Function
* `image_size` - Image size for Yandex Cloud Function

* `connectivity` - Function version connectivity. If specified the version will be attached to specified network
* `connectivity.0.network_id` - Network the version will have access to. It's essential to specify network with subnets in all availability zones

* `package` - Version deployment package for Yandex Cloud Function code. Can be only one `package` or `content` section. Either `package` or `content` section must be specified
* `package.0.sha_256` - SHA256 hash of the version deployment package
* `package.0.bucket_name` - Name of the bucket that stores the code for the version
* `package.0.object_name` - Name of the object in the bucket that stores the code for the version

* `content` - Version deployment content for Yandex Cloud Function code. Can be only one `package` or `content` section. Either `package` or `content` section must be specified
* `content.0.zip_filename` - Filename to zip archive for the version

* `async_invocation` - Config for asynchronous invocations of Yandex Cloud Function
* `log_options` - Options for logging from Yandex Cloud Function
* `tmpfs_size` - Tmpfs size for Yandex Cloud Function
* `concurrency` - The maximum number of requests processed by a function instance at the same time

* `metadata_options` - Options set the access mode to function's metadata endpoints.
* `metadata_options.0.gce_http_endpoint` - Enables access to GCE flavored metadata. Values: `0`- default, `1` - enabled, `2` - disabled
* `metadata_options.0.aws_v1_http_endpoint` Enables access to AWS flavored metadata (IMDSv1). Values: `0` - default, `1` - enabled, `2` - disabled

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the Yandex Cloud Function
* `version` - Version for Yandex Cloud Function
* `image_size` - Image size for Yandex Cloud Function

---

The `secrets` block supports:

* `id` - (Required) Secret's id
* `version_id` - (Required) Secret's version id
* `key` - (Required) Secret's entries key which value will be stored in environment variable
* `environment_variable` - (Required) Function's environment variable in which secret's value will be stored. Must begin with a letter (A-Z, a-z).

---

The `async_invocation` block supports:

* `retries_count` - Maximum number of retries for async invocation
* `service_account_id` - Service account used for async invocation
* `ymq_success_target` - Target for successful async invocation
* `ymq_failure_target` - Target for unsuccessful async invocation

---

Both `ymq_success_target` and `ymq_failure_target` blocks supports:

* `arn` - YMQ ARN
* `service_account_id` - Service account used for writing result to queue

---

The `log_options` block supports:

* `disabled` - Is logging from function disabled
* `log_group_id` - Log entries are written to specified log group
* `folder_id` - Log entries are written to default log group for specified folder
* `min_level` - Minimum log entry level

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_function.<resource Name> <resource Id>
terraform import yandex_function.test-function d4e45**********pqvd3
```
