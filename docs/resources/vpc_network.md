---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_network"
description: |-
  Manages a network within Yandex Cloud.
---

# yandex_vpc_network (Resource)

Manages a network within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/network#network).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)
  * [VPC Addressing](https://yandex.cloud/docs/vpc/concepts/address)

## Example usage

```terraform
//
// Create a new VPC Network.
//
resource "yandex_vpc_network" "my_net" {
  name = "foobar"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the network. Provided by the client when the network is created.

* `description` - (Optional) An optional description of this resource. Provide this property when you create the resource.

* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to apply to this network. A list of key/value pairs.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `default_security_group_id` - ID of default Security Group of this network.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_vpc_network.<resource Name> <resource Id>
terraform import yandex_vpc_network.my_net ...
```
