---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a VPC Private Endpoint within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a VPC Private Endpoint within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/private-endpoint).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)

## Example usage

{{ tffile "examples/vpc_private_endpoint/r_vpc_private_endpoint_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the private endpoint. Provided by the client when the private endpoint is created.
* `description` - (Optional) An optional description of this resource. Provide this property when you create the resource.
* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.
* `labels` - (Optional) Labels to apply to this resource. A list of key/value pairs.
* `network_id` - (Required) ID of the network which private endpoint belongs to.
* `endpoint_address` - (Optional) Private endpoint address specification block.
* `dns_options` - (Optional) Private endpoint DNS options block.

---

The `endpoint_address` block supports:
* `address_id` - ID of the address.
* `subnet_id` - Subnet of the IP address.
* `address` - Specifies IP address within `subnet_id`.

~> Only one of `address_id` or `subnet_id` + `address` arguments can be specified.

---

The `dns_options` block supports:
* `private_dns_records_enabled` - If enabled - additional service dns will be created.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - Status of the private endpoint.
* `created_at` - Creation timestamp of the key.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/vpc_private_endpoint/import.sh" }}

