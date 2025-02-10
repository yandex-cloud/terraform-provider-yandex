---
subcategory: "Yandex API Gateway"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex API Gateway.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud API Gateway](https://yandex.cloud/docs/api-gateway/).

## Example usage

{{ tffile "examples/api_gateway/r_api_gateway_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` (Required) - Yandex Cloud API Gateway name used to define API Gateway.
* `spec` - (Required) OpenAPI specification for Yandex API Gateway.
* `folder_id` - (Optional) Folder ID for the Yandex Cloud API Gateway. If it is not provided, the default provider folder is used.
* `description` - (Optional) Description of the Yandex Cloud API Gateway.
* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Cloud API Gateway.
* `custom_domains` - (Optional) Set of custom domains to be attached to Yandex API Gateway.
* `connectivity` - (Optional) Gateway connectivity. If specified the gateway will be attached to specified network.
* `connectivity.0.network_id` - Network the gateway will have access to. It's essential to specify network with subnets in all availability zones.
* `variables` - (Optional) A set of values for variables in gateway specification.
* `canary` - (Optional) Canary release settings of gateway.
* `canary.0.weight` - Percentage of requests, which will be processed by canary release.
* `canary.0.variables` - A list of values for variables in gateway specification of canary release.
* `log_options` - Options for logging from Yandex Cloud API Gateway.
* `execution_timeout` - Execution timeout in seconds for the Yandex Cloud API Gateway.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the Yandex Cloud API Gateway.
* `domain` - Default domain for the Yandex API Gateway. Generated at creation time.
* `loggroup_id` - ID of the log group for the Yandex API Gateway.
* `status` - Status of the Yandex API Gateway.
* `user_domains` - (**DEPRECATED**, use `custom_domains` instead) Set of user domains attached to Yandex API Gateway.

---

* The `log_options` block supports:
* `disabled` - Is logging from API Gateway disabled
* `log_group_id` - Log entries are written to specified log group
* `folder_id` - Log entries are written to default log group for specified folder
* `min_level` - Minimum log entry level

## Import

~> Import for this resource is not implemented yet.
