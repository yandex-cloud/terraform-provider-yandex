---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about Yandex Application Load Balancer Backend Group.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Application Load Balancer Backend Group. For more information, see [official documentation](https://yandex.cloud/docs/application-load-balancer/quickstart).

## Example usage

{{ tffile "examples/alb_backend_group/d_alb_backend_group_1.tf" }}

This data source is used to define [Application Load Balancer Backend Groups](https://yandex.cloud/docs/application-load-balancer/concepts/backend-group) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `backend_group_id` (Optional) - Backend Group ID.

* `name` - (Optional) - Name of the Backend Group.

~> One of `backend_group_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the backend group.
* `labels` - Labels to assign to this backend group.
* `session_affinity` - Session affinity mode determines how incoming requests are grouped into one session. Structure is documented below.
* `http_backend` - Http backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `grpc_backend` - Grpc backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `stream_backend` - Stream backend specification that will be used by the ALB Backend Group. Structure is documented below.
* `created_at` - Creation timestamp of this backend group.

~> Only one type of backends `http_backend` or `grpc_backend` or `stream_backend` should be specified.

The `session_affinity` block supports:

* `connection` - Requests received from the same IP are combined into a session. Stream backend groups only support session affinity by client IP address. Structure is documented below.
* `cookie` - Requests with the same cookie value and the specified file name are combined into a session. Allowed only for HTTP and gRPC backend groups. Structure is documented below.
* `header` - Requests with the same value of the specified HTTP header, such as with user authentication data, are combined into a session. Allowed only for HTTP and gRPC backend groups. Structure is documented below.

~> Only one type( `connection` or `cookie` or `header` ) of session affinity should be specified.

The `connection` block supports:

* `source_ip` - Source IP address to use with affinity.

The `cookie` block supports:

* `name` - Name of the HTTP cookie to use with affinity.
* `ttl` - TTL for the cookie (if not set, session cookie will be used)

The `header` block supports:

* `header_name` - The name of the request header that will be used with affinity.

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
* `keep_connections_on_host_health_failure` - If set, when a backend host becomes unhealthy (as determined by the configured health checks), keep connections to the failed host.

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

~> Only one of `validation_context.0.trusted_ca_id` or `validation_context.0.trusted_ca_bytes` should be specified.

The `load_balancing_config` block supports:

* `panic_threshold` - If percentage of healthy hosts in the backend is lower than panic_threshold, traffic will be routed to all backends no matter what the health status is. This helps to avoid healthy backends overloading when everything is bad. Zero means no panic threshold.
* `locality_aware_routing_percent` - Percent of traffic to be sent to the same availability zone. The rest will be equally divided between other zones.
* `strict_locality` - If set, will route requests only to the same availability zone. Balancer won't know about endpoints in other zones.
* `mode` - Load balancing mode for the backend.

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

~> Only one of `stream_healthcheck` or `http_healthcheck` or `grpc_healthcheck` should be specified.

The `stream_healthcheck` block supports:

* `send` - Optional message text sent to targets during TCP data transfer.
* `receive` - Optional text that must be contained in the messages received from targets for a successful health check.

The `http_healthcheck` block supports:

* `host` - Optional "Host" HTTP header value.
* `path` - HTTP path.
* `http2` - If set, health checks will use HTTP2.
* `expected_statuses` - (Optional) A list of HTTP response statuses considered healthy.

The `grpc_healthcheck` block supports:

* `service_name` - Optional service name for grpc.health.v1.HealthCheckRequest message.
