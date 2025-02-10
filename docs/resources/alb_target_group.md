---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: yandex_alb_target_group"
description: |-
  An application load balancer distributes the load across cloud resources that are combined into a target group.
---

# yandex_alb_target_group (Resource)

Creates a target group in the specified folder and adds the specified targets to it. For more information, see [the official documentation](https://yandex.cloud/docs/application-load-balancer/concepts/target-group).

## Example usage

```terraform
//
// Create a new ALB Target Group
//
resource "yandex_alb_target_group" "foo" {
  name = "my-target-group"

  target {
    subnet_id  = yandex_vpc_subnet.my-subnet.id
    ip_address = yandex_compute_instance.my-instance-1.network_interface.0.ip_address
  }

  target {
    subnet_id  = yandex_vpc_subnet.my-subnet.id
    ip_address = yandex_compute_instance.my-instance-2.network_interface.0.ip_address
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the target group. Provided by the client when the target group is created.

* `description` - (Optional) An optional description of the target group. Provide this property when you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this target group. A list of key/value pairs.

* `target` - (Optional) A Target resource. The structure is documented below.

---

The `target` block supports:

* `ip_address` - (Required) IP address of the target.

* `subnet_id` - (Required) ID of the subnet that targets are connected to. All targets in the target group must be connected to the same subnet within a single availability zone.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the target group.

* `created_at` - The target group creation timestamp.

## Timeouts

This resource provides the following configuration options for timeouts:

- `create` - Default is 5 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_alb_target_group.<resource Name> <resource Id>
terraform import yandex_alb_target_group.my_target_group ds77d**********g4pqc
```
