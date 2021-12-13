---
layout: "yandex"
page_title: "Yandex: yandex_alb_virtual_host"
sidebar_current: "docs-yandex-datasource-alb-virtual-host"
description: |- Get information about a Yandex ALB Virtual Host.
---

# yandex\_alb\_virtual\_host

Get information about a Yandex ALB Virtual Host. For more information, see
[Yandex.Cloud Application Load Balancer](https://cloud.yandex.com/en/docs/application-load-balancer/quickstart).

## Example Usage

```hcl
data "yandex_alb_virtual_host" "my-vh-data" {
  name = yandex_alb_virtual_host.my-vh.name
  http_router_id = yandex_alb_virtual_host.my-router.id
}
```

This data source is used to define [Application Load Balancer Virtual Host] that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `virtual_host_id` - (Optional) The ID of a specific Virtual Host. Virtual Host ID is concatenation of HTTP Router ID
  and Virtual Host name with `/` symbol between them. For Example, "http_router_id/vhost_name".
* `name` - (Optional) Name of the Virtual Host.
* `http_router_id` - (Optional) HTTP Router that the resource belongs to.

~> **NOTE:** One of `virtual_host_id` or `name` with `http_router_id` should be specified.

## Attributes Reference

The following attributes are exported:

* `authority` - A list of domains (host/authority header) that will be matched to this virtual host. Wildcard hosts are
  supported in the form of '*.foo.com' or '*-bar.foo.com'. If not specified, all domains will be matched.

* `modify_request_headers` - Apply the following modifications to the request headers. The structure is documented
  below.

* `modify_response_headers` - Apply the following modifications to the response headers. The structure is documented
  below.

* `route` - A Route resource. Routes are matched *in-order*. Be careful when adding them to the end. For instance,
  having http '/' match first makes all other routes unused. The structure is documented below.

---

The `modify_request_headers` and `modify_response_headers` blocks support:

* `name` - name of the header to modify.

* `append` - Append string to the header value.

* `replace` - New value for a header. Header values support the following
  [formatters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#custom-request-response-headers).

* `remove` - If set, remove the header.

~> **NOTE:** Only one type of actions `append` or `replace` or `remove` should be specified.

---

The `route` block supports:

* `name` - name of the route.

* `http_route` - HTTP route resource. The structure is documented below.

* `grpc_route` - GRPC route resource. The structure is documented below.

~> **NOTE:** Exactly one type of routes `http_route` or `grpc_route` should be specified.

---

The `http_route` block supports:

* `http_match` - Checks "/" prefix by default. The structure is documented below.

* `http_route_action` - HTTP route action resource. The structure is documented below.

* `redirect_action` - Redirect action resource. The structure is documented below.

* `direct_response_action` - (Required) Direct response action resource. The structure is documented below.

~> **NOTE:** Exactly one type of actions `http_route_action` or `redirect_action` or `direct_response_action` should be
specified.

---

The `http_match` block supports:

* `http_method` - List of methods(strings).

* `path` - If not set, '/' is assumed. The structure is documented below.

---

The `http_route_action` block supports:

* `backend_group_id` - Backend group to route requests.

* `timeout` - Specifies the request timeout (overall time request processing is allowed to take) for the route. If not
  set, default is 60 seconds.

* `idle_timeout` - Specifies the idle timeout (time without any data transfer for the active request) for the route. It
  is useful for streaming scenarios (i.e. long-polling, server-sent events) - one should set idle_timeout to something
  meaningful and timeout to the maximum time the stream is allowed to be alive. If not specified, there is no per-route
  idle timeout.

* `host_rewrite` - Host rewrite specifier.

* `auto_host_rewrite` - If set, will automatically rewrite host.

* `prefix_rewrite` - If not empty, matched path prefix will be replaced by this value.

* `upgrade_types` - List of upgrade types. Only specified upgrade types will be allowed. For example,
  "websocket".

~> **NOTE:** Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be specified.

---

The `direct_response_action` block supports:

* `status` - HTTP response status. Should be between 100 and 599.

* `body` - Response body text.

---

The `redirect_action` block supports:

* `replace_scheme` - Replaces scheme. If the original scheme is `http` or `https`, will also remove the 80 or 443 port,
  if present.

* `replace_host` - Replaces hostname.

* `replace_port` - Replaces port.

* `replace_path` - Replace path.

* `replace_prefix` - Replace only matched prefix. Example:<br/> match:{ prefix_match: "/some" } <br/>
  redirect: { replace_prefix: "/other" } <br/> will redirect "/something" to "/otherthing".

* `remove query` - If set, remove query part.

* `response_code` - The HTTP status code to use in the redirect response. Supported values are:
  moved_permanently, found, see_other, temporary_redirect, permanent_redirect.

~> **NOTE:** Only one type of paths `replace_path` or `replace_prefix` should be specified.

---

The `grpc_route` block supports:

* `grpc_match` - Checks "/" prefix by default. The structure is documented below.

* `grpc_route_action` - GRPC route action resource. The structure is documented below.

* `grpc_status_response_action` - (Required) GRPC status response action resource. The structure is documented below.

~> **NOTE:** Exactly one type of actions `grpc_route_action` or `grpc_status_response_action` should be specified.

---

The `grpc_match` block supports:

* `fqmn` - If not set, all services/methods are assumed. The structure is documented below.

---

The `grpc_route_action` block supports:

* `backend_group_id` - Backend group to route requests.

* `max_timeout` - Lower timeout may be specified by the client (using grpc-timeout header). If not set, default is 60
  seconds.

* `idle_timeout` - Specifies the idle timeout (time without any data transfer for the active request) for the route. It
  is useful for streaming scenarios - one should set idle_timeout to something meaningful and max_timeout to the maximum
  time the stream is allowed to be alive. If not specified, there is no per-route idle timeout.

* `host_rewrite` - Host rewrite specifier.

* `auto_host_rewrite` - If set, will automatically rewrite host.

~> **NOTE:** Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be specified.

---

The `grpc_status_response_action` block supports:

* `status` - The status of the response. Supported values are: ok, invalid_argumet, not_found, permission_denied,
  unauthenticated, unimplemented, internal, unavailable.

---

The `path` and `fqmn` blocks support:

* `exact_match` - Match exactly.

* `prefix_match` - Match prefix.

~> **NOTE:** Exactly one type of string matches `exact_match` or `prefix_match` should be specified.


[Application Load Balancer Virtual Host]: https://cloud.yandex.com/en/docs/application-load-balancer/concepts/http-router