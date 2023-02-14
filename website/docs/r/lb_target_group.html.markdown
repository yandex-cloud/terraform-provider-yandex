---
layout: "yandex"
page_title: "Yandex: yandex_lb_target_group"
sidebar_current: "docs-yandex-lb-target-group"
description: |-
  A load balancer distributes the load across cloud resources that are combined into a target group.
---

Creates a target group in the specified folder and adds the specified targets to it. 
For more information, see [the official documentation](https://cloud.yandex.com/docs/load-balancer/concepts/target-resources).

# yandex\_lb\_target\_group

## Example Usage

```hcl
resource "yandex_lb_target_group" "foo" {
  name      = "my-target-group"
  region_id = "ru-central1"

  target {
    subnet_id = "${yandex_vpc_subnet.my-subnet.id}"
    address   = "${yandex_compute_instance.my-instance-1.network_interface.0.ip_address}"
  }

  target {
    subnet_id = "${yandex_vpc_subnet.my-subnet.id}"
    address   = "${yandex_compute_instance.my-instance-2.network_interface.0.ip_address}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the target group. Provided by the client when the target group is created.

* `description` - (Optional) An optional description of the target group. Provide this property when
you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs.
If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this target group. A list of key/value pairs.

* `region_id` - (Optional) ID of the availability zone where the target group resides.
If omitted, default region is being used.

* `target` - (Optional) A Target resource. The structure is documented below.

---

The `target` block supports:

* `address` - (Required) IP address of the target.

* `subnet_id` - (Required) ID of the subnet that targets are connected to. 
All targets in the target group must be connected to the same subnet within a single availability zone. 

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the target group.

* `created_at` - The target group creation timestamp.

## Timeouts

This resource provides the following configuration options for 
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 5 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.

## Import

A target group can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_lb_target_group.default target_group_id
```
