---
layout: "yandex"
page_title: "Yandex: yandex_alb_virtual_host"
sidebar_current: "docs-yandex-alb-virtual-host"
description: |- Virtual hosts combine routes belonging to the same set of domains.
---

Creates a virtual host that belongs to specified HTTP router and adds the specified routes to it. For more information,
see [the official documentation](https://cloud.yandex.com/en/docs/application-load-balancer/concepts/http-router).

# yandex\_alb\_virtual\_host

## Example Usage

```hcl
resource "yandex_alb_virtual_host" "my-virtual-host" {
  name      = "my-virtual-host"
  http_router_id = yandex_alb_http_router.my-router.id
  route {
    name = "my-route"
    http_route {
      http_route_action {
        backend_group_id = yandex_alb_backend_group.my-bg.id
        timeout = "3s"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the virtual host. Provided by the client when the virtual host is created.

* `"http_router_id` - (Required) The ID of the HTTP router to which the virtual host belongs.

* `authority` - (Optional) A list of domains (host/authority header) that will be matched to this virtual host. Wildcard
  hosts are supported in the form of '*.foo.com' or '*-bar.foo.com'. If not specified, all domains will be matched.

* `modify_request_headers` - (Optional) Apply the following modifications to the request
  headers. The structure is documented below.

* `modify_response_headers` - (Optional) Apply the following modifications to the response
  headers. The structure is documented below.

* `route` - (Optional) A Route resource. Routes are matched *in-order*. Be careful when adding them to the end. For instance, having
  http '/' match first makes all other routes unused. The structure is documented below.

---

The `modify_request_headers` and `modify_response_headers` blocks support:

* `name` - (Required) name of the header to modify.

* `append` - (Optional) Append string to the header value.

* `replace` - (Optional) New value for a header. Header values support the following 
  [formatters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#custom-request-response-headers).

* `remove` - (Optional) If set, remove the header.

~> **NOTE:** Only one type of actions `append` or `replace` or `remove` should be specified.

---

The `route` block supports:

* `name` - (Required) name of the route.

* `http_route` - (Optional) HTTP route resource. The structure is documented below.

* `grpc_route` - (Optional) GRPC route resource. The structure is documented below.

~> **NOTE:** Exactly one type of routes `http_route` or `grpc_route` should be specified.

---

The `http_route` block supports:

* `http_match` - (Optional) Checks "/" prefix by default. The structure is documented below.

* `http_route_action` - (Optional) HTTP route action resource. The structure is documented below.

* `redirect_action` - (Optional) Redirect action resource. The structure is documented below.

* `direct_response_action` - (Required) Direct response action resource. The structure is documented below.

~> **NOTE:** Exactly one type of actions `http_route_action` or `redirect_action` or `direct_response_action` should be 
specified.

---

The `http_match` block supports:

* `http_method` - (Optional) List of methods(strings).

* `path` - (Optional) If not set, '/' is assumed. The structure is documented below.

---

The `http_route_action` block supports:

* `backend_group_id` - (Required) Backend group to route requests.

* `timeout` - (Optional) Specifies the request timeout (overall time request processing is allowed to take) for the 
  route. If not set, default is 60 seconds.

* `idle_timeout` - (Optional) Specifies the idle timeout (time without any data transfer for the active request) for the 
  route. It is useful for streaming scenarios (i.e. long-polling, server-sent events) - one should set idle_timeout to 
  something meaningful and timeout to the maximum time the stream is allowed to be alive. If not specified, there is no 
  per-route idle timeout.

* `host_rewrite` - (Optional) Host rewrite specifier.

* `auto_host_rewrite` - (Optional) If set, will automatically rewrite host.

* `prefix_rewrite` - (Optional) If not empty, matched path prefix will be replaced by this value.

* `upgrade_types` - (Optional) List of upgrade types. Only specified upgrade types will be allowed. For example, 
  "websocket".

~> **NOTE:** Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be
specified.

---

The `direct_response_action` block supports:

* `status` - (Optional) HTTP response status. Should be between 100 and 599.

* `body` - (Optional) Response body text.

---

The `redirect_action` block supports:

* `replace_scheme` - (Optional) Replaces scheme. If the original scheme is `http` or `https`, will also remove the 
  80 or 443 port, if present.
  
* `replace_host` - (Optional) Replaces hostname.

* `replace_port` - (Optional) Replaces port.
  
* `replace_path` - (Optional) Replace path.

* `replace_prefix` - (Optional) Replace only matched prefix. Example:<br/> match:{ prefix_match: "/some" } <br/> 
  redirect: { replace_prefix: "/other" } <br/> will redirect "/something" to "/otherthing".

* `remove query` - (Optional) If set, remove query part.

* `response_code` - (Optional) The HTTP status code to use in the redirect response. Supported values are: 
moved_permanently, found, see_other, temporary_redirect, permanent_redirect.

~> **NOTE:** Only one type of paths `replace_path` or `replace_prefix` should be specified.

---

The `grpc_route` block supports:

* `grpc_match` - (Optional) Checks "/" prefix by default. The structure is documented below.

* `grpc_route_action` - (Optional) GRPC route action resource. The structure is documented below.

* `grpc_status_response_action` - (Required) GRPC status response action resource. The structure is documented below.

~> **NOTE:** Exactly one type of actions `grpc_route_action` or `grpc_status_response_action` should be specified.

---

The `grpc_match` block supports:

* `fqmn` - (Optional) If not set, all services/methods are assumed. The structure is documented below.

---

The `grpc_route_action` block supports:

* `backend_group_id` - (Required) Backend group to route requests.

* `max_timeout` - (Optional) Lower timeout may be specified by the client (using grpc-timeout header). If not set, default is 
  60 seconds.

* `idle_timeout` - (Optional) Specifies the idle timeout (time without any data transfer for the active request) for the
  route. It is useful for streaming scenarios - one should set idle_timeout to something meaningful and max_timeout 
  to the maximum time the stream is allowed to be alive. If not specified, there is no
  per-route idle timeout.

* `host_rewrite` - (Optional) Host rewrite specifier.

* `auto_host_rewrite` - (Optional) If set, will automatically rewrite host.

~> **NOTE:** Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be
specified.

---

The `grpc_status_response_action` block supports:

* `status` - (Optional) The status of the response. Supported values are: ok, invalid_argumet, not_found, 
  permission_denied, unauthenticated, unimplemented, internal, unavailable. 

---

The `path` and `fqmn` blocks support:

* `exact` - (Optional) Match exactly.
  
* `prefix` - (Optional) Match prefix.

~> **NOTE:** Exactly one type of string matches `exact` or `prefix` should be
specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the virtual host.

## Timeouts

This resource provides the following configuration options for timeouts:

- `create` - Default is 5 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.

## Import

A virtual host can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_alb_virtual_host.default virtual_host_id
```