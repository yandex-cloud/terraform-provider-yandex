---
subcategory: "Serverless Cloud Functions"
page_title: "Yandex: yandex_function"
description: |-
  Get information about a Yandex Cloud Function.
---

# yandex_function (Data Source)

Get information about a Yandex Cloud Function. For more information about Yandex Cloud Functions, see [Yandex Cloud Functions](https://yandex.cloud/docs/functions).
This data source is used to define [Yandex Cloud Function](https://yandex.cloud/docs/functions/concepts/function) that can be used by other resources.

~> Either `function_id` or `name` must be specified.

## Example usage

```terraform
//
// Get information about existing Yandex Cloud Function
//
data "yandex_function" "my_function" {
  function_id = "d4e45**********pqvd3"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `concurrency` (Number) The maximum number of requests processed by a function instance at the same time.
- `connectivity` (Block List, Max: 1) (see [below for nested schema](#nestedblock--connectivity))
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `function_id` (String) Yandex Cloud Function id used to define function.
- `metadata_options` (Block List, Max: 1) (see [below for nested schema](#nestedblock--metadata_options))
- `mounts` (Block List) (see [below for nested schema](#nestedblock--mounts))
- `name` (String) The resource name.
- `secrets` (Block List) (see [below for nested schema](#nestedblock--secrets))
- `storage_mounts` (Block List, Deprecated) (see [below for nested schema](#nestedblock--storage_mounts))

### Read-Only

- `async_invocation` (List of Object) (see [below for nested schema](#nestedatt--async_invocation))
- `created_at` (String) The creation timestamp of the resource.
- `description` (String) The resource description.
- `entrypoint` (String) Entrypoint for Yandex Cloud Function.
- `environment` (Map of String) A set of key/value environment variables for Yandex Cloud Function. Each key must begin with a letter (A-Z, a-z).
- `execution_timeout` (String) Execution timeout in seconds for Yandex Cloud Function.
- `id` (String) The ID of this resource.
- `image_size` (Number) Image size for Yandex Cloud Function.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `log_options` (List of Object) (see [below for nested schema](#nestedatt--log_options))
- `memory` (Number) Memory in megabytes (**aligned to 128MB**) for Yandex Cloud Function.
- `runtime` (String) Runtime for Yandex Cloud Function.
- `service_account_id` (String) [Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) which linked to the resource.
- `tags` (Set of String) Tags for Yandex Cloud Function. Tag `$latest` isn't returned.
- `tmpfs_size` (Number) Tmpfs size for Yandex Cloud Function.
- `version` (String) Version of Yandex Cloud Function.

<a id="nestedblock--connectivity"></a>
### Nested Schema for `connectivity`

Required:

- `network_id` (String) Network the version will have access to. It's essential to specify network with subnets in all availability zones.



<a id="nestedblock--metadata_options"></a>
### Nested Schema for `metadata_options`

Optional:

- `aws_v1_http_endpoint` (Number) Enables access to AWS flavored metadata (IMDSv1). Values: `0` - default, `1` - enabled, `2` - disabled.

- `gce_http_endpoint` (Number) Enables access to GCE flavored metadata. Values: `0`- default, `1` - enabled, `2` - disabled.



<a id="nestedblock--mounts"></a>
### Nested Schema for `mounts`

Required:

- `name` (String) Name of the mount point. The directory where the target is mounted will be accessible at the `/function/storage/<mounts.0.name>` path.


Optional:

- `ephemeral_disk` (Block List, Max: 1) One of the available mount types. Disk available during the function execution time. (see [below for nested schema](#nestedblock--mounts--ephemeral_disk))

- `mode` (String) Mount’s accessibility mode. Valid values are `ro` and `rw`.

- `object_storage` (Block List, Max: 1) One of the available mount types. Object storage as a mount. (see [below for nested schema](#nestedblock--mounts--object_storage))


<a id="nestedblock--mounts--ephemeral_disk"></a>
### Nested Schema for `mounts.ephemeral_disk`

Required:

- `size_gb` (Number) Size of the ephemeral disk in GB.


Optional:

- `block_size_kb` (Number) Optional block size of the ephemeral disk in KB.



<a id="nestedblock--mounts--object_storage"></a>
### Nested Schema for `mounts.object_storage`

Required:

- `bucket` (String) Name of the mounting bucket.


Optional:

- `prefix` (String) Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted.




<a id="nestedblock--secrets"></a>
### Nested Schema for `secrets`

Required:

- `environment_variable` (String) Function's environment variable in which secret's value will be stored. Must begin with a letter (A-Z, a-z).

- `id` (String) Secret's ID.

- `key` (String) Secret's entries key which value will be stored in environment variable.

- `version_id` (String) Secret's version ID.



<a id="nestedblock--storage_mounts"></a>
### Nested Schema for `storage_mounts`

Required:

- `bucket` (String) Name of the mounting bucket.

- `mount_point_name` (String) Name of the mount point. The directory where the bucket is mounted will be accessible at the `/function/storage/<mount_point>` path.


Optional:

- `prefix` (String) Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted.

- `read_only` (Boolean) Mount the bucket in read-only mode.



<a id="nestedatt--async_invocation"></a>
### Nested Schema for `async_invocation`

Read-Only:

- `retries_count` (Number) Maximum number of retries for async invocation.

- `service_account_id` (String) Service account used for async invocation.

- `ymq_failure_target` (Block List, Max: 1) Target for unsuccessful async invocation. (see [below for nested schema](#nestedobjatt--async_invocation--ymq_failure_target))

- `ymq_success_target` (Block List, Max: 1) Target for successful async invocation. (see [below for nested schema](#nestedobjatt--async_invocation--ymq_success_target))


<a id="nestedobjatt--async_invocation--ymq_failure_target"></a>
### Nested Schema for `async_invocation.ymq_failure_target`

Read-Only:

- `arn` (String) YMQ ARN.

- `service_account_id` (String) Service account used for writing result to queue.



<a id="nestedobjatt--async_invocation--ymq_success_target"></a>
### Nested Schema for `async_invocation.ymq_success_target`

Read-Only:

- `arn` (String) YMQ ARN.

- `service_account_id` (String) Service account used for writing result to queue.




<a id="nestedatt--log_options"></a>
### Nested Schema for `log_options`

Read-Only:

- `disabled` (Boolean) Is logging from function disabled.

- `folder_id` (String) Log entries are written to default log group for specified folder.

- `log_group_id` (String) Log entries are written to specified log group.

- `min_level` (String) Minimum log entry level.

