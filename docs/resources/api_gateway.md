---
subcategory: "Serverless API Gateway"
page_title: "Yandex: yandex_api_gateway"
description: |-
  Allows management of a Yandex Cloud API Gateway.
---


# yandex_api_gateway




Allows management of [Yandex Cloud API Gateway](https://cloud.yandex.com/docs/api-gateway/).

## Example usage

```terraform
resource "yandex_api_gateway" "test-api-gateway" {
  name        = "some_name"
  description = "any description"
  labels = {
    label       = "label"
    empty-label = ""
  }
  custom_domains {
    fqdn           = "test.example.com"
    certificate_id = "<certificate_id_from_cert_manager>"
  }
  connectivity {
    network_id = "<dynamic network id>"
  }
  variables = {
    installation = "prod"
  }
  canary {
    weight = 20
    variables = {
      installation = "dev"
    }
  }
  log_options {
    log_group_id = "<log group id>"
    min_level    = "ERROR"
  }
  execution_timeout = "300"
  spec              = <<-EOT
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test API
x-yc-apigateway:
  variables:
    installation:
      default: "prod"
      enum:
       - "prod"
       - "dev"
paths:
  /hello:
    get:
      summary: Say hello
      operationId: hello
      parameters:
        - name: user
          in: query
          description: User name to appear in greetings
          required: false
          schema:
            type: string
            default: 'world'
      responses:
        '200':
          description: Greeting
          content:
            'text/plain':
              schema:
                type: "string"
      x-yc-apigateway-integration:
        type: dummy
        http_code: 200
        http_headers:
          'Content-Type': "text/plain"
        content:
          'text/plain': "Hello again, {user} from ${apigw.installation} release!\n"
EOT
}
```

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
