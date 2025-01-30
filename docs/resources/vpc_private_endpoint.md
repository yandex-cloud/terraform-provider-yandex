---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_private_endpoint"
description: |-
  Manages a VPC Private Endpoint within Yandex Cloud.
---

# yandex_vpc_private_endpoint (Resource)

Manages a VPC Private Endpoint within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/private-endpoint).

* How-to Guides
  * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)

## Example usage

```terraform
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_subnet" "lab-subnet-a" {
  v4_cidr_blocks = ["10.2.0.0/16"]
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.lab-net.id
}

resource "yandex_vpc_private_endpoint" "default" {
  name        = "object-storage-private-endpoint"
  description = "description for private endpoint"

  labels = {
    my-label = "my-label-value"
  }

  network_id = yandex_vpc_network.lab-net.id

  object_storage {}

  dns_options {
    private_dns_records_enabled = true
  }

  endpoint_address {
    subnet_id = yandex_vpc_subnet.lab-subnet-a.id
  }
}
```

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

Private endpoint can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_vpc_private_endpoint.pe private_endpoint_id
```
