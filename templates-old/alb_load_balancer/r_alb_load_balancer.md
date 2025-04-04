---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: {{.Name}}"
description: |-
  A Load Balancer is used for receiving incoming traffic and transmitting it to the backend endpoints
  specified in the ALB Target Groups.
---

# {{.Name}} ({{.Type}})

Creates an Application Load Balancer in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/application-load-balancer/concepts/application-load-balancer) .

## Example usage

{{ tffile "examples/alb_load_balancer/r_alb_load_balancer_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the Load Balancer. Provided by the client when the Load Balancer is created.

* `description` - (Optional) An optional description of the Load Balancer.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this Load Balancer. A list of key/value pairs.

* `region_id` - (Optional) ID of the region that the Load Balancer is located at.

* `network_id` - (Required) ID of the network that the Load Balancer is located at.

* `security_group_ids` - (Optional) A list of ID's of security groups attached to the Load Balancer.

* `allocation_policy` - (Required) Allocation zones for the Load Balancer instance. The structure is documented below.

* `listener` - (Optional) List of listeners for the Load Balancer. The structure is documented below.

* `log_options` - (Optional) Cloud Logging settings. The structure is documented below.

---

The `allocation_policy` block supports:

* `location` - (Required) Unique set of locations. The structure is documented below.

---

The `location` block supports:

* `zone_id` - (Required) ID of the zone that location is located at.

* `subnet_id` - (Required) ID of the subnet that location is located at.

* `disable_traffic` - (Optional) If set, will disable all L7 instances in the zone for request handling.

---

The `listener` block supports:

* `name` - (Required) name of the listener.

* `endpoint` - (Required) Network endpoints (addresses and ports) of the listener. The structure is documented below.

* `http` - (Optional) HTTP listener resource. The structure is documented below.

* `stream` - (Optional) Stream listener resource. The structure is documented below.

* `tls` - (Optional) TLS listener resource. The structure is documented below.

~> Exactly one listener type: `http` or `tls` or `stream` should be specified.

---

The `endpoint` block supports:

* `address` - (Required) One or more addresses to listen on. The structure is documented below.

* `ports` - (Required) One or more ports to listen on.

---

The `address` block supports:

* `external_ipv4_address` - (Optional) External IPv4 address. The structure is documented below.

* `internal_ipv4_address` - (Optional) Internal IPv4 address. The structure is documented below.

* `external_ipv6_address` - (Optional) External IPv6 address. The structure is documented below.

~> Exactly one type of addresses `external_ipv4_address` or `internal_ipv4_address` or `external_ipv6_address` should be specified.

---

The `external_ipv4_address` block supports:

* `address` - (Optional) Provided by the client or computed automatically.

---

The `internal_ipv4_address` block supports:

* `address` - (Optional) Provided by the client or computed automatically.

* `subnet_id` - (Required) Provided by the client or computed automatically.

---

The `external_ipv6_address` block supports:

* `address` - (Optional) Provided by the client or computed automatically.

---

The `tls` block supports:

* `default_handler` - (Required) TLS handler resource. The structure is documented below.

* `sni_handler` - (Optional) SNI match resource. The structure is documented below.

---

The `sni_handler` block supports:

* `name` - (Required) name of SNI match.

* `server_names` - (Required) A set of server names.

* `handler` - (Required) TLS handler resource. The structure is documented below.

---

The `default_handler` and `handler`(from `sni_handler`) block supports:

* `http_handler` - (Required) HTTP handler resource. The structure is documented below.

* `stream_handler` - (Required) Stream handler resource. The structure is documented below.

* `certificate_ids` - (Required) Certificate IDs in the Certificate Manager. Multiple TLS certificates can be associated with the same context to allow both RSA and ECDSA certificates. Only the first certificate of each type will be used.

~> Exactly one handler type `http_handler` or `stream_handler` should be specified.

---

The `stream` block supports:

* `handler` - (Optional) Stream handler that sets plaintext Stream backend group. The structure is documented below.

---

The `stream_handler` and `handler`(from `stream`) block supports:

* `backend_group_id` - (Optional) Backend group id.

* `idle_timeout` - The idle timeout is the interval after which the connection will be forcibly closed if no data has been transmitted or received on either the upstream or downstream connection. If not configured, the default idle timeout is 1 hour. Setting it to 0 disables the timeout.

---

The `http` block supports:

* `handler` - (Optional) HTTP handler that sets plaintext HTTP router. The structure is documented below.

* `redirects` - (Optional) Shortcut for adding http -> https redirects. The structure is documented below.

~> Only one type of fields `handler` or `redirects` should be specified.

---

The `http_handler` and `handler`(from `http`) block supports:

* `http_router_id` - (Optional) HTTP router id.

* `rewrite_request_id` - (Optional) When unset, will preserve the incoming x-request-id header, otherwise would rewrite it with a new value.

* `http2_options` - (Optional) If set, will enable HTTP2 protocol for the handler. The structure is documented below.

* `allow_http10` - (Optional) If set, will enable only HTTP1 protocol with HTTP1.0 support.

~> Only one type of protocol settings `http2_options` or `allow_http10` should be specified.

---

The `http2_options` block supports:

* `max_concurrent_streams` - (Optional) Maximum number of concurrent streams.

---

The `redirects` block supports:

* `http_to_https` - (Optional) If set redirects all unencrypted HTTP requests to the same URI with scheme changed to `https`.

---

The `log_options` block supports:

* `disable` (Optional) Set to true to disable Cloud Logging for the balancer

* `log_group_id` (Optional) Cloud Logging group ID to send logs to. Leave empty to use the balancer folder default log group.

* `discard_rule` (Optional) List of rules to discard a fraction of logs. The structure is documented below.

---

The `discard_rule` block supports:

* `http_codes` (Optional) list of http codes *100*-*599*

* `http_code_intervals` (Optional) list of http code intervals *1XX*-*5XX* or *ALL*

* `grpc_codes` (Optional) list of grpc codes by name, e.g, *["NOT_FOUND", "RESOURCE_EXHAUSTED"]*

---

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the Load Balancer.

* `created_at` - The Load Balancer creation timestamp.

* `status` - Status of the Load Balancer.

* `log_group_id` - Cloud log group used by the Load Balancer to store access logs.

## Timeouts

This resource provides the following configuration options for timeouts:

- `create` - Default is 10 minute.
- `update` - Default is 10 minute.
- `delete` - Default is 10 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/alb_load_balancer/import.sh" }}
