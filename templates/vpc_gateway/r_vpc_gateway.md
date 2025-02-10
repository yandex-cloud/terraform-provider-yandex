---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a gateway within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a gateway within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/gateways).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)

## Example usage

{{ tffile "examples/vpc_gateway/r_vpc_gateway_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the VPC Gateway. Provided by the client when the VPC Gateway is created.

* `description` - (Optional) An optional description of this resource. Provide this property when you create the resource.

* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to apply to this VPC Gateway. A list of key/value pairs.

* `shared_egress_gateway` - Shared egress gateway configuration. Currently empty.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/vpc_gateway/import.sh" }}
