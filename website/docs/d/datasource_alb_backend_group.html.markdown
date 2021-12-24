---
layout: "yandex"
page_title: "Yandex: yandex_alb_backend_group"
sidebar_current: "docs-yandex-datasource-alb-backend-group"
description: |-
Get information about a Yandex Application Load Balancer Backend Group.
---

# yandex\_alb\_backend\_group

Get information about a Yandex Application Load Balancer Backend Group. For more information, see
[Yandex.Cloud Application Load Balancer](https://cloud.yandex.com/en/docs/application-load-balancer/quickstart).

```hcl
data "yandex_alb_backend_group" "foo" {
  backend_group_id = "my-backend-group-id"
}
```

This data source is used to define [Application Load Balancer Backend Groups] that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `backend_group_id` (Optional) - Backend Group ID.

* `name` - (Optional) - Name of the Backend Group.

~> **NOTE:** One of `backend_group_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the backend group.
* `labels` - Labels to assign to this backend group.
* `http_backend` - Http backend specification that will be used by the ALB Backend Group. Structure is documented below. 
* `grpc_backend` - Grpc backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `stream_backend` - Stream backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `created_at` - Creation timestamp of this backend group.

~> **NOTE:** Only one type of backends `http_backend` or `grpc_backend` or `stream_backend` should be specified.

The `http_backend` block supports:

* `name` - Name of the backend.
* `port` - Port for incoming traffic.
* `weight` - Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights. 
* `http2` - Enables HTTP2 for upstream requests. If not set, HTTP 1.1 will be used by default.
* `target_group_ids` - References target groups for the backend.
* `load_balancing_config` - Load Balancing Config specification that will be used by this backend. Structure is documented below.
* `healthcheck` - Healthcheck specification that will be used by this backend. Structure is documented below.
* `tls` - Tls specification that will be used by this backend. Structure is documented below.

The `stream_backend` block supports:

* `name` - Name of the backend.
* `port` - Port for incoming traffic.
* `weight` - Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.
* `target_group_ids` - References target groups for the backend.
* `load_balancing_config` - Load Balancing Config specification that will be used by this backend. Structure is documented below.
* `healthcheck` - Healthcheck specification that will be used by this backend. Structure is documented below.
* `tls` - Tls specification that will be used by this backend. Structure is documented below.

The `grpc_backend` block supports:

* `name` - Name of the backend.
* `port` - Port for incoming traffic.
* `weight` - Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.
* `target_group_ids` - References target groups for the backend.
* `load_balancing_config` - Load Balancing Config specification that will be used by this backend. Structure is documented below.
* `healthcheck` - Healthcheck specification that will be used by this backend. Structure is documented below.
* `tls` - Tls specification that will be used by this backend. Structure is documented below.

The `tls` block supports:

* `sni` - [SNI](https://en.wikipedia.org/wiki/Server_Name_Indication) string for TLS connections.
* `validation_context.0.trusted_ca_id` - Trusted CA certificate ID in the Certificate Manager.
* `validation_context.0.trusted_ca_bytes` - PEM-encoded trusted CA certificate chain.

~> **NOTE:** Only one of `validation_context.0.trusted_ca_id` or `validation_context.0.trusted_ca_bytes` should be specified.

The `load_balancing_config` block supports:

* `panic_threshold` - If percentage of healthy hosts in the backend is lower than panic_threshold, traffic will be routed to all backends no matter what the health status is. This helps to avoid healthy backends overloading  when everything is bad. Zero means no panic threshold.
* `locality_aware_routing_percent` - Percent of traffic to be sent to the same availability zone. The rest will be equally divided between other zones.
* `strict_locality` - If set, will route requests only to the same availability zone. Balancer won't know about endpoints in other zones.

The `healthcheck` block supports:

* `timeout` - Time to wait for a health check response.
* `interval` - Interval between health checks.
* `interval_jitter_percent` - An optional jitter amount as a percentage of interval. If specified, during every interval value of (interval_ms * interval_jitter_percent / 100) will be added to the wait time.
* `healthy_threshold` - Number of consecutive successful health checks required to promote endpoint into the healthy state. 0 means 1. Note that during startup, only a single successful health check is required to mark a host healthy.
* `unhealthy_threshold` - Number of consecutive failed health checks required to demote endpoint into the unhealthy state. 0 means 1. Note that for HTTP health checks, a single 503 immediately makes endpoint unhealthy.
* `healthcheck_port` - Optional alternative port for health checking.
* `stream_healthcheck` - Stream Healthcheck specification that will be used by this healthcheck. Structure is documented below.
* `http_healthcheck` - Http Healthcheck specification that will be used by this healthcheck. Structure is documented below.
* `grpc_healthcheck` - Grpc Healthcheck specification that will be used by this healthcheck. Structure is documented below.

~> **NOTE:** Only one of `stream_healthcheck` or `http_healthcheck` or `grpc_healthcheck` should be specified.

The `stream_healthcheck` block supports:

* `send` - Optional message to send. If empty, it's a connect-only health check.
* `receive` - Optional text to search in reply.

The `http_healthcheck` block supports:

* `host` - Optional "Host" HTTP header value.
* `path` - HTTP path.
* `http2` - If set, health checks will use HTTP2.

The `grpc_healthcheck` block supports:

* `service_name` - Optional service name for grpc.health.v1.HealthCheckRequest message.

[Application Load Balancer Backend Groups]: https://cloud.yandex.com/en/docs/application-load-balancer/concepts/backend-group