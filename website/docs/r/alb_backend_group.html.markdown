---
layout: "yandex"
page_title: "Yandex: yandex_alb_backend_group"
sidebar_current: "docs-yandex-alb-backend-group"
description: |-
An application load balancer distributes the load across cloud resources that are combined into a backend group.
---

Creates a backend group in the specified folder and adds the specified backends to it.
For more information, see [the official documentation](https://cloud.yandex.com/en/docs/application-load-balancer/concepts/backend-group).

# yandex\_alb\_backend\_group

## Example Usage

```hcl
resource "yandex_alb_backend_group" "test-backend-group" {
  name      = "my-backend-group"

  http_backend {
    name = "test-http-backend"
    weight = 1
    port = 8080
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    tls {
      sni = "backend-domain.internal"
    }
    load_balancing_config {
      panic_threshold = 50
    }    
    healthcheck {
      timeout = "1s"
      interval = "1s"
      http_healthcheck {
        path  = "/"
      }
    }
    http2 = "true"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the Backend Group.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `description` - (Optional) Description of the backend group.
* `labels` - (Optional) Labels to assign to this backend group.
* `http_backend` - (Optional) Http backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `grpc_backend` - (Optional) Grpc backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `stream_backend` - (Optional) Stream backend specification that will be used by the ALB Backend Group. Structure is documented below.

~> **NOTE:** Only one type of backends `http_backend` or `grpc_backend` or `stream_backend` should be specified.

The `http_backend` block supports:

* `name` - (Required) Name of the backend.
* `port` - (Optional) Port for incoming traffic.
* `weight` - (Optional) Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.
* `http2` - (Optional) Enables HTTP2 for upstream requests. If not set, HTTP 1.1 will be used by default.
* `target_group_ids` - (Required) References target groups for the backend.
* `load_balancing_config` - (Optional) Load Balancing Config specification that will be used by this backend. Structure is documented below.
* `healthcheck` - (Optional) Healthcheck specification that will be used by this backend. Structure is documented below.
* `tls` - (Optional) Tls specification that will be used by this backend. Structure is documented below.

The `stream_backend` block supports:

* `name` - (Required) Name of the backend.
* `port` - (Optional) Port for incoming traffic.
* `weight` - (Optional) Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.
* `target_group_ids` - (Required) References target groups for the backend.
* `load_balancing_config` - (Optional) Load Balancing Config specification that will be used by this backend. Structure is documented below.
* `healthcheck` - (Optional) Healthcheck specification that will be used by this backend. Structure is documented below.
* `tls` - (Optional) Tls specification that will be used by this backend. Structure is documented below.

The `grpc_backend` block supports:

* `name` - (Required) Name of the backend.
* `port` - (Optional) Port for incoming traffic.
* `weight` - (Optional) Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.
* `target_group_ids` - (Required) References target groups for the backend.
* `load_balancing_config` - (Optional) Load Balancing Config specification that will be used by this backend. Structure is documented below.
* `healthcheck` - (Optional) Healthcheck specification that will be used by this backend. Structure is documented below.
* `tls` - (Optional) Tls specification that will be used by this backend. Structure is documented below.

The `tls` block supports:

* `sni` - (Optional) [SNI](https://en.wikipedia.org/wiki/Server_Name_Indication) string for TLS connections.
* `validation_context.0.trusted_ca_id` - (Optional) Trusted CA certificate ID in the Certificate Manager.
* `validation_context.0.trusted_ca_bytes` - (Optional) PEM-encoded trusted CA certificate chain.

~> **NOTE:** Only one of `validation_context.0.trusted_ca_id` or `validation_context.0.trusted_ca_bytes` should be specified.

The `load_balancing_config` block supports:

* `panic_threshold` - (Optional) If percentage of healthy hosts in the backend is lower than panic_threshold, traffic will be routed to all backends no matter what the health status is. This helps to avoid healthy backends overloading  when everything is bad. Zero means no panic threshold.
* `locality_aware_routing_percent` - (Optional) Percent of traffic to be sent to the same availability zone. The rest will be equally divided between other zones.
* `strict_locality` - (Optional) If set, will route requests only to the same availability zone. Balancer won't know about endpoints in other zones.
* `mode` - (Optional) Load balancing mode for the backend. Possible values: "ROUND_ROBIN", "RANDOM", "LEAST_REQUEST", "MAGLEV_HASH".

The `healthcheck` block supports:

* `timeout` - (Required) Time to wait for a health check response.
* `interval` - (Required) Interval between health checks.
* `interval_jitter_percent` - (Optional) An optional jitter amount as a percentage of interval. If specified, during every interval value of (interval_ms * interval_jitter_percent / 100) will be added to the wait time.
* `healthy_threshold` - (Optional) Number of consecutive successful health checks required to promote endpoint into the healthy state. 0 means 1. Note that during startup, only a single successful health check is required to mark a host healthy.
* `unhealthy_threshold` - (Optional) Number of consecutive failed health checks required to demote endpoint into the unhealthy state. 0 means 1. Note that for HTTP health checks, a single 503 immediately makes endpoint unhealthy.
* `healthcheck_port` - (Optional) Optional alternative port for health checking.
* `stream_healthcheck` - (Optional) Stream Healthcheck specification that will be used by this healthcheck. Structure is documented below.
* `http_healthcheck` - (Optional) Http Healthcheck specification that will be used by this healthcheck. Structure is documented below.
* `grpc_healthcheck` - (Optional) Grpc Healthcheck specification that will be used by this healthcheck. Structure is documented below.

~> **NOTE:** Only one of `stream_healthcheck` or `http_healthcheck` or `grpc_healthcheck` should be specified.

The `stream_healthcheck` block supports:

* `send` - (Optional) Message sent to targets during TCP data transfer.  If not specified, no data is sent to the target.
* `receive` - (Optional) Data that must be contained in the messages received from targets for a successful health check. If not specified, no messages are expected from targets, and those that are received are not checked.

The `http_healthcheck` block supports:

* `host` - (Optional) "Host" HTTP header value.
* `path` - (Required) HTTP path.
* `http2` - (Optional) If set, health checks will use HTTP2.

The `grpc_healthcheck` block supports:

* `service_name` - (Optional) Service name for grpc.health.v1.HealthCheckRequest message.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the backend group.

* `created_at` - The backend group creation timestamp.

## Timeouts

This resource provides the following configuration options for
timeouts:

- `create` - Default is 5 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.

## Import

A backend group can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_alb_backend_group.default backend_group_id
```