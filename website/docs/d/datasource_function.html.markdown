---
layout: "yandex"
page_title: "Yandex: yandex_function"
sidebar_current: "docs-yandex-datasource-yandex-function"
description: |-
  Get information about a Yandex Cloud Function.
---

# yandex\_function

Get information about a Yandex Cloud Function. For more information about Yandex Cloud Functions, see 
[Yandex Cloud Functions](https://cloud.yandex.com/docs/functions/).

```hcl
data "yandex_function" "my_function" {
  function_id = "are1samplefunction11"
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
* `tags` - Tags for Yandex Cloud Function. Tag "$latest" isn't returned.
* `secrets` - Secrets for Yandex Cloud Function. 
* `version` - Version for Yandex Cloud Function.
* `image_size` - Image size for Yandex Cloud Function.
* `loggroup_id` - Log group ID size for Yandex Cloud Function.
* `connectivity` - Function version connectivity. If specified the version will be attached to specified network.
* `connectivity.0.network_id` - Network the version will have access to. It's essential to specify network with subnets in all availability zones.




