---
subcategory: "Network Load Balancer (NLB)"
page_title: "Yandex: {{.Name}}"
description: |-
  A network load balancer is used to evenly distribute the load across cloud resources.
---

# {{.Name}} ({{.Type}})

Creates a network load balancer in the specified folder using the data specified in the config. For more information, see [the official documentation](https://yandex.cloud/docs/load-balancer/concepts).

## Example usage

{{ tffile "examples/lb_network_load_balancer/r_lb_network_load_balancer_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the network load balancer. Provided by the client when the network load balancer is created.

* `description` - (Optional) An optional description of the network load balancer. Provide this property when you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this network load balancer. A list of key/value pairs.

* `region_id` - (Optional) ID of the availability zone where the network load balancer resides. If omitted, default region is being used.

* `type` - (Optional) Type of the network load balancer. Must be one of 'external' or 'internal'. The default is 'external'.

* `attached_target_group` - (Optional) An AttachedTargetGroup resource. The structure is documented below.

* `listener` - (Optional) Listener specification that will be used by a network load balancer. The structure is documented below.

* `deletion_protection` - (Optional) Flag that protects the network load balancer from accidental deletion.

* `allow_zonal_shift` - (Optional) Flag that marks the network load balancer as avaialable to zonal shift.

---

The `attached_target_group` block supports:

* `target_group_id` - (Required) ID of the target group.

* `healthcheck` - (Required) A HealthCheck resource. The structure is documented below.

---

The `healthcheck` block supports:

* `name` - (Required) Name of the health check. The name must be unique for each target group that attached to a single load balancer.

* `interval` - (Optional) The interval between health checks. The default is 2 seconds.

* `timeout` - (Optional) Timeout for a target to return a response for the health check. The default is 1 second.

* `unhealthy_threshold` - (Optional) Number of failed health checks before changing the status to `UNHEALTHY`. The default is 2.

* `healthy_threshold` - (Optional) Number of successful health checks required in order to set the `HEALTHY` status for the target.

* `http_options` - (Optional) Options for HTTP health check. The structure is documented below.

* `tcp_options` - (Optional) Options for TCP health check. The structure is documented below.

~> One of `http_options` or `tcp_options` should be specified.

---

The `http_options` block supports:

* `port` - (Required) Port to use for HTTP health checks.

* `path` - (Optional) URL path to set for health checking requests for every target in the target group. For example `/ping`. The default path is `/`.

---

The `tcp_options` block supports:

* `port` - (Required) Port to use for TCP health checks.

---

The `listener` block supports:

* `name` - (Required) Name of the listener. The name must be unique for each listener on a single load balancer.

* `port` - (Required) Port for incoming traffic.

* `target_port` - (Optional) Port of a target. The default is the same as listener's port.

* `protocol` - (Optional) Protocol for incoming traffic. TCP or UDP and the default is TCP.

* `external_address_spec` - (Optional) External IP address specification. The structure is documented below.

* `internal_address_spec` - (Optional) Internal IP address specification. The structure is documented below.

~> One of `external_address_spec` or `internal_address_spec` should be specified.

---

The `external_address_spec` block supports:

* `address` - (Optional) External IP address for a listener. IP address will be allocated if it wasn't been set.

* `ip_version` - (Optional) IP version of the external addresses that the load balancer works with. Must be one of ipv4 or ipv6. The default is ipv4.

---

The `internal_address_spec` block supports:

* `subnet_id` - (Required) ID of the subnet to which the internal IP address belongs.

* `address` - (Optional) Internal IP address for a listener. Must belong to the subnet that is referenced in subnet_id. IP address will be allocated if it wasn't been set.

* `ip_version` - (Optional) IP version of the internal addresses that the load balancer works with. Must be one of ipv4 or ipv6. The default is ipv4.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the network load balancer.

* `created_at` - The network load balancer creation timestamp.

## Timeouts

This resource provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 5 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/lb_network_load_balancer/import.sh" }}
