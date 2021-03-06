---
layout: "yandex"
page_title: "Yandex: yandex_api_gateway"
sidebar_current: "docs-yandex-datasource-yandex-api-gateway"
description: |-
  Get information about a Yandex Cloud API Gateway.
---

# yandex\_api\_gateway

Get information about a Yandex Cloud API Gateway. For more information, see the official documentation 
[Yandex Cloud API Gateway](https://cloud.yandex.com/docs/api-gateway/).

```hcl
data "yandex_api_gateway" "my-api-gateway" {
  name = "my-api-gateway"
}
```

## Argument Reference

The following arguments are supported:

* `api_gateway_id` (Optional) - Yandex Cloud API Gateway id used to define api gateway.

* `name` (Optional) - Yandex Cloud API Gateway name used to define api gateway.

* `folder_id` (Optional) - Folder ID for the Yandex Cloud API Gateway.

~> **NOTE:** Either `api_gateway_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the Yandex Cloud API Gateway.
* `labels` - A set of key/value label pairs to assign to the Yandex Cloud API Gateway.
* `created_at` - Creation timestamp of the Yandex Cloud API Gateway.
* `loggroup_id` - ID of the log group for the Yandex API Gateway.
* `domain` - Default domain for the Yandex API Gateway.
* `status` - Status of the Yandex API Gateway.
* `user_domains` - Set of user domains attached to Yandex API Gateway.