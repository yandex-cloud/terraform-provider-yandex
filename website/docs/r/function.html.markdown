---
layout: "yandex"
page_title: "Yandex: yandex_function"
sidebar_current: "docs-yandex-function"
description: |-
 Allows management of a Yandex Cloud Function.
---

# yandex\_function

Allows management of [Yandex Cloud Function](https://cloud.yandex.com/docs/functions/)

## Example Usage

```hcl
resource "yandex_function" "test-function" {
  name               = "some_name"
  description        = "any description"
  user_hash          = "any_user_defined_string"
  runtime            = "python37"
  entrypoint         = "main"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = "are1service2account3id"
  tags               = ["my_tag"]
  secrets {
    id = "${yandex_lockbox_secret.secret.id}"
    version_id = "${yandex_lockbox_secret_version.secret_version.id}"
    key = "secret-key"
    environment_variable = "ENV_VARIABLE"
  }
  content {
    zip_filename = "function.zip"
  }
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
* `environment` - A set of key/value environment variables for Yandex Cloud Function
* `tags` - Tags for Yandex Cloud Function. Tag "$latest" isn't returned.
* `secrets` - Secrets for Yandex Cloud Function.
* `version` - Version for Yandex Cloud Function.
* `image_size` - Image size for Yandex Cloud Function.
* `loggroup_id` - Loggroup ID size for Yandex Cloud Function.

* `connectivity` - Function version connectivity. If specified the version will be attached to specified network.
* `connectivity.0.network_id` - Network the version will have access to. It's essential to specify network with subnets in all availability zones.

* `package` - Version deployment package for Yandex Cloud Function code. Can be only one `package` or `content` section. Either `package` or `content` section must be specified.
* `package.0.sha_256` - SHA256 hash of the version deployment package.
* `package.0.bucket_name` - Name of the bucket that stores the code for the version.
* `package.0.object_name` - Name of the object in the bucket that stores the code for the version.

* `content` - Version deployment content for Yandex Cloud Function code. Can be only one `package` or `content` section. Either `package` or `content` section must be specified.
* `content.0.zip_filename` - Filename to zip archive for the version. 


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the Yandex Cloud Function.
* `version` - Version for Yandex Cloud Function.
* `image_size` - Image size for Yandex Cloud Function.
* `loggroup_id` - Log group ID size for Yandex Cloud Function.

---

The `secrets` block supports:

* `id` - (Required) Secret's id.

* `version_id` - (Required) Secret's version id.

* `key` - (Required) Secret's entries key which value will be stored in environment variable.

* `environment_variable` - (Required) Function's environment variable in which secret's value will be stored.
