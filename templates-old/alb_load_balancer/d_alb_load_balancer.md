---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Application Load Balancer.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Application Load Balancer. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).

## Example usage

{{ tffile "examples/alb_load_balancer/d_alb_load_balancer_1.tf" }}

This data source is used to define [Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/concepts/application-load-balancer) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - Load Balancer ID.

* `name` - Name of the Load Balancer.

~> One of `load_balancer_id` or `name` should be specified.

* `folder_id` - Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - An optional description of the Load Balancer.

* `folder_id` - The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - Labels to assign to this Load Balancer. A list of key/value pairs.

* `region_id` - ID of the region that the Load Balancer is located at.

* `network_id` - ID of the network that the Load Balancer is located at.

* `security_group_ids` - A list of ID's of security groups attached to the Load Balancer.

* `allocation_policy` - Allocation zones for the Load Balancer instance. The structure is documented below.

* `listener` - List of listeners for the Load Balancer. The structure is documented below.

* `log_options` - Cloud Logging settings. The structure is documented below.

* `created_at` - The Load Balancer creation timestamp.

* `status` - Status of the Load Balancer.

* `log_group_id` - Cloud log group used by the Load Balancer to store access logs.

---

The `allocation_policy` block supports:

* `location` - Unique set of locations. The structure is documented below.

---

The `location` block supports:

* `zone_id` - ID of the zone that location is located at.

* `subnet_id` - ID of the subnet that location is located at.

* `disable_traffic` - If set, will disable all L7 instances in the zone for request handling.

---

The `listener` block supports:

* `name` - name of the listener.

* `endpoint` - Network endpoints (addresses and ports) of the listener. The structure is documented below.

* `http` - HTTP listener resource. The structure is documented below.

* `tls` - TLS listener resource. The structure is documented below.

* `stream` - Stream listener resource. The structure is documented below.

~> Exactly one listener type: `http` or `tls` or `stream` should be specified.

---

The `endpoint` block supports:

* `address` - One or more addresses to listen on. The structure is documented below.

* `ports` - One or more ports to listen on.

---

The `address` block supports:

* `external_ipv4_address` - External IPv4 address. The structure is documented below.

* `internal_ipv4_address` - Internal IPv4 address. The structure is documented below.

* `external_ipv6_address` - External IPv6 address. The structure is documented below.

~> Exactly one type of addresses `external_ipv4_address` or `internal_ipv4_address` or `external_ipv6_address` should be specified.

---

The `external_ipv4_address` block supports:

* `address` - Provided by the client or computed automatically.

---

The `internal_ipv4_address` block supports:

* `address` - Provided by the client or computed automatically.

* `subnet_id` - Provided by the client or computed automatically.

---

The `external_ipv6_address` block supports:

* `address` - Provided by the client or computed automatically.

---

The `tls` block supports:

* `default_handler` - TLS handler resource. The structure is documented below.

* `sni_handler` - SNI match resource. The structure is documented below.

---

The `sni_handler` block supports:

* `name` - name of SNI match.

* `server_names` - A set of server names.

* `handler` - TLS handler resource. The structure is documented below.

---

The `default_handler` and `handler`(from `sni_handler`) block supports:

* `http_handler` - HTTP handler resource. The structure is documented below.

* `stream_handler` - Stream handler resource. The structure is documented below.

* `certificate_ids` - Certificate IDs in the Certificate Manager. Multiple TLS certificates can be associated with the same context to allow both RSA and ECDSA certificates. Only the first certificate of each type will be used.

~> Exactly one handler type `http_handler` or `stream_handler` should be specified.

---

The `stream` block supports:

* `handler` - Stream handler that sets plaintext Stream backend group. The structure is documented below.

---

The `http` block supports:

* `handler` - HTTP handler that sets plaintext HTTP router. The structure is documented below.

* `redirects` - Shortcut for adding http -> https redirects. The structure is documented below.

~> Only one type of fields `handler` or `redirects` should be specified.

---

The `http_handler` and `handler`(from `http`) block supports:

* `http_router_id` - HTTP router id.

* `rewrite_request_id` - When unset, will preserve the incoming x-request-id header, otherwise would rewrite it with a new value.

* `http2_options` - If set, will enable HTTP2 protocol for the handler. The structure is documented below.

* `allow_http10` - If set, will enable only HTTP1 protocol with HTTP1.0 support.

~> Only one type of protocol settings `http2_options` or `allow_http10` should be specified.

---

The `stream_handler` and `handler`(from `stream`) block supports:

* `backend_group_id` - Backend group id.

* `idle_timeout` - The idle timeout is the interval after which the connection will be forcibly closed if no data has been transmitted or received on either the upstream or downstream connection. If not configured, the default idle timeout is 1 hour. Setting it to 0 disables the timeout.

---

The `http2_options` block supports:

* `max_concurrent_streams` - Maximum number of concurrent streams.

The `redirects` block supports:

* `http_to_https` - If set redirects all unencrypted HTTP requests to the same URI with scheme changed to `https`.

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
