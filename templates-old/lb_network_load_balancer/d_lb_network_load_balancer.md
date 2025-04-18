---
subcategory: "Network Load Balancer (NLB)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Load Balancer network load balancer.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Load Balancer network load balancer. For more information, see [the official documentation](https://yandex.cloud/docs/load-balancer/concepts/).

## Example usage

{{ tffile "examples/lb_network_load_balancer/d_lb_network_load_balancer_1.tf" }}

This data source is used to define [Load Balancer Network Load Balancers](https://yandex.cloud/docs/load-balancer/concepts/) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `network_load_balancer_id` (Optional) - Network load balancer ID.

* `name` - (Optional) - Name of the network load balancer.

~> One of `network_load_balancer_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the network load balancer.
* `labels` - Labels to assign to this network load balancer.
* `region_id` - ID of the region where the network load balancer resides.
* `type` - Type of the network load balancer.
* `attached_target_group` - An attached target group is a group of targets that is attached to a load balancer. Structure is documented below.
* `listener` - Listener specification that will be used by a network load balancer. Structure is documented below.
* `created_at` - Creation timestamp of this network load balancer.
* `deletion_protection` - Flag that protects the network load balancer from accidental deletion.
* `allow_zonal_shift` -  Flag that marks the network load balancer as avaialable to zonal shift.

---

The `attached_target_group` block supports:

* `target_group_id` - ID of the target group that attached to the network load balancer.
* `healthcheck.0.name` - Name of the health check.
* `healthcheck.0.interval` - The interval between health checks.
* `healthcheck.0.timeout` - Timeout for a target to return a response for the health check.
* `healthcheck.0.unhealthy_threshold` - Number of failed health checks before changing the status to `UNHEALTHY`.
* `healthcheck.0.healthy_threshold` - Number of successful health checks required in order to set the `HEALTHY` status for the target.
* `healthcheck.0.tcp_options.0.port` - Port to use for TCP health checks.
* `healthcheck.0.http_options.0.port` - Port to use for HTTP health checks.
* `healthcheck.0.http_options.0.path` - URL path to use for HTTP health checks.

The `listener` block supports:

* `name` - Name of the listener.
* `port` - Port for incoming traffic.
* `protocol` - Protocol for incoming traffic.
* `target_port` - Port of a target.
* `external_address_spec.0.address` - External IP address of a listener.
* `external_address_spec.0.ip_version` - IP version of the external addresses.
* `internal_address_spec.0.subnet_id` - Subnet ID to which the internal IP address belongs
* `internal_address_spec.0.address` - Internal IP address of a listener.
* `internal_address_spec.0.ip_version` - IP version of the internal addresses.
