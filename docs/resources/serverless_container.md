---
subcategory: "Serverless Containers"
page_title: "Yandex: yandex_serverless_container"
description: |-
  Allows management of a Yandex Cloud Serverless Container.
---

# yandex_serverless_container (Resource)

Allows management of Yandex Cloud Serverless Containers.

## Example usage

```terraform
resource "yandex_serverless_container" "test-container" {
  name               = "some_name"
  description        = "any description"
  memory             = 256
  execution_timeout  = "15s"
  cores              = 1
  core_fraction      = 100
  service_account_id = "are1service2account3id"
  runtime {
    type = "task"
  }
  secrets {
    id                   = yandex_lockbox_secret.secret.id
    version_id           = yandex_lockbox_secret_version.secret_version.id
    key                  = "secret-key"
    environment_variable = "ENV_VARIABLE"
  }
  mounts {
    mount_point_path = "/mount/point"
    ephemeral_disk {
      size_gb = 5
    }
  }
  image {
    url = "cr.yandex/yc/test-image:v1"
  }
  log_options {
    log_group_id = "e2392vo6d1bne2aeq9fr"
    min_level    = "ERROR"
  }
  provision_policy {
    min_instances = 1
  }
}
```

### Serverless Container with Image Digest

```terraform
resource "yandex_serverless_container" "test-container-with-digest" {
  name   = "some_name"
  memory = 128
  image {
    url    = "cr.yandex/yc/test-image:v1"
    digest = "sha256:e1d772fa8795adac847a2420c87d0d2e3d38fb02f168cab8c0b5fe2fb95c47f4"
  }
}
```

### Serverless Container with Mounted Object Storage Bucket

```terraform
locals {
  folder_id = "folder_id"
}

resource "yandex_serverless_container" "test-container-object-storage-mount" {
  name               = "some_name"
  memory             = 128
  service_account_id = yandex_iam_service_account.sa.id
  image {
    url = "cr.yandex/yc/test-image:v1"
  }
  mounts {
    mount_point_path = "/mount/point"
    mode             = "ro"
    object_storage {
      bucket = yandex_storage_bucket.my-bucket.bucket
    }
  }
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
* `runtime` - Runtime for Yandex Cloud Serverless Container
* `runtime.0.type` - Type of the runtime for Yandex Cloud Serverless Container. Valid values are `http` and `task`

* `secrets` - Secrets for Yandex Cloud Serverless Container

* `storage_mounts` - (**DEPRECATED**, use `mounts.0.object_storage` instead) Storage mounts for Yandex Cloud Serverless Container
* `storage_mounts.0.mount_point_path` - (Required) Path inside the container to access the directory in which the bucket is mounted
* `storage_mounts.0.bucket` - (Required) Name of the mounting bucket
* `storage_mounts.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted
* `storage_mounts.0.read_only` - Mount the bucket in read-only mode

* `mounts` - Mounts for Yandex Cloud Serverless Container
* `mounts.0.mount_point_path` - (Required) Path inside the container to access the directory in which the target is mounted
* `mounts.0.mode` - Mount’s accessibility mode. Valid values are `ro` and `rw`
* `mounts.0.ephemeral_disk` - One of the available mount types. Disk available during the function execution time
* `mounts.0.ephemeral_disk.0.size_gb` - (Required) Size of the ephemeral disk in GB
* `mounts.0.ephemeral_disk.0.block_size_kb` - Optional block size of the ephemeral disk in KB
* `mounts.0.object_storage` - One of the available mount types. Object storage as a mount
* `mounts.0.object_storage.0.bucket` - (Required) Name of the mounting bucket
* `mounts.0.object_storage.0.prefix` - Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted

* `connectivity` - Network access. If specified the revision will be attached to specified network
* `connectivity.0.network_id` - Network the revision will have access to

* `image` - Revision deployment image for Yandex Cloud Serverless Container
* `image.0.url` (Required) - URL of image that will be deployed as Yandex Cloud Serverless Container
* `image.0.work_dir` - Working directory for Yandex Cloud Serverless Container
* `image.0.digest` - Digest of image that will be deployed as Yandex Cloud Serverless Container. If presented, should be equal to digest that will be resolved at server side by URL. Container will be updated on digest change even if `image.0.url` stays the same. If field not specified then its value will be computed.
* `image.0.command` - List of commands for Yandex Cloud Serverless Container
* `image.0.args` - List of arguments for Yandex Cloud Serverless Container
* `image.0.environment` - A set of key/value environment variable pairs for Yandex Cloud Serverless Container. Each key must begin with a letter (A-Z, a-z).

* `log_options` - Options for logging from Yandex Cloud Serverless Container

* `provision_policy` - Provision policy. If specified the revision will have prepared instances
* `provision_policy.0.min_instances` - Minimum number of prepared instances that are always ready to serve requests

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `url` - Invoke URL for the Yandex Cloud Serverless Container
* `created_at` - Creation timestamp of the Yandex Cloud Serverless Container
* `revision_id` - Last revision ID of the Yandex Cloud Serverless Container

---

The `secrets` block supports:

* `id` - (Required) Secret's id
* `version_id` - (Required) Secret's version id
* `key` - (Required) Secret's entries key which value will be stored in environment variable
* `environment_variable` - (Required) Container's environment variable in which secret's value will be stored. Must begin with a letter (A-Z, a-z).

---

* The `log_options` block supports:
* `disabled` - Is logging from container disabled
* `log_group_id` - Log entries are written to specified log group
* `folder_id` - Log entries are written to default log group for specified folder
* `min_level` - Minimum log entry level
