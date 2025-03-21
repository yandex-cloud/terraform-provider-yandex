---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: yandex_alb_load_balancer"
description: |-
  Get information about a Yandex Application Load Balancer.
---

# yandex_alb_load_balancer (Data Source)

Get information about a Yandex Application Load Balancer. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).

This data source is used to define [Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/concepts/application-load-balancer) that can be used by other resources.

~> One of `load_balancer_id` or `name` should be specified.

## Example usage

```terraform
//
// Get information about existing Application Load Balancer (ALB).
//
data "yandex_alb_load_balancer" "tf-alb-data" {
  load_balancer_id = "my-alb-id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `load_balancer_id` (String) The resource identifier.
- `name` (String) The resource name.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `allocation_policy` (List of Object) Allocation zones for the Load Balancer instance. (see [below for nested schema](#nestedatt--allocation_policy))
- `created_at` (String) The creation timestamp of the resource.
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `id` (String) The ID of this resource.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `listener` (List of Object) List of listeners for the Load Balancer. (see [below for nested schema](#nestedatt--listener))
- `log_group_id` (String) Cloud Logging group ID to send logs to. Leave empty to use the balancer folder default log group.
- `log_options` (List of Object) Cloud Logging settings. (see [below for nested schema](#nestedatt--log_options))
- `network_id` (String) The `VPC Network ID` of subnets which resource attached to.
- `region_id` (String) The region ID where Load Balancer is located at.
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `status` (String) Status of the Load Balancer.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedatt--allocation_policy"></a>
### Nested Schema for `allocation_policy`

Read-Only:

- `location` (Set of Object) (see [below for nested schema](#nestedobjatt--allocation_policy--location))

<a id="nestedobjatt--allocation_policy--location"></a>
### Nested Schema for `allocation_policy.location`

Read-Only:

- `disable_traffic` (Boolean)
- `subnet_id` (String)
- `zone_id` (String)



<a id="nestedatt--listener"></a>
### Nested Schema for `listener`

Read-Only:

- `endpoint` (List of Object) (see [below for nested schema](#nestedobjatt--listener--endpoint))
- `http` (List of Object) (see [below for nested schema](#nestedobjatt--listener--http))
- `name` (String)
- `stream` (List of Object) (see [below for nested schema](#nestedobjatt--listener--stream))
- `tls` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls))

<a id="nestedobjatt--listener--endpoint"></a>
### Nested Schema for `listener.endpoint`

Read-Only:

- `address` (List of Object) (see [below for nested schema](#nestedobjatt--listener--endpoint--address))
- `ports` (List of Number)

<a id="nestedobjatt--listener--endpoint--address"></a>
### Nested Schema for `listener.endpoint.address`

Read-Only:

- `external_ipv4_address` (List of Object) (see [below for nested schema](#nestedobjatt--listener--endpoint--address--external_ipv4_address))
- `external_ipv6_address` (List of Object) (see [below for nested schema](#nestedobjatt--listener--endpoint--address--external_ipv6_address))
- `internal_ipv4_address` (List of Object) (see [below for nested schema](#nestedobjatt--listener--endpoint--address--internal_ipv4_address))

<a id="nestedobjatt--listener--endpoint--address--external_ipv4_address"></a>
### Nested Schema for `listener.endpoint.address.external_ipv4_address`

Read-Only:

- `address` (String)


<a id="nestedobjatt--listener--endpoint--address--external_ipv6_address"></a>
### Nested Schema for `listener.endpoint.address.external_ipv6_address`

Read-Only:

- `address` (String)


<a id="nestedobjatt--listener--endpoint--address--internal_ipv4_address"></a>
### Nested Schema for `listener.endpoint.address.internal_ipv4_address`

Read-Only:

- `address` (String)
- `subnet_id` (String)




<a id="nestedobjatt--listener--http"></a>
### Nested Schema for `listener.http`

Read-Only:

- `handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--http--handler))
- `redirects` (List of Object) (see [below for nested schema](#nestedobjatt--listener--http--redirects))

<a id="nestedobjatt--listener--http--handler"></a>
### Nested Schema for `listener.http.handler`

Read-Only:

- `allow_http10` (Boolean)
- `http2_options` (List of Object) (see [below for nested schema](#nestedobjatt--listener--http--handler--http2_options))
- `http_router_id` (String)
- `rewrite_request_id` (Boolean)

<a id="nestedobjatt--listener--http--handler--http2_options"></a>
### Nested Schema for `listener.http.handler.http2_options`

Read-Only:

- `max_concurrent_streams` (Number)



<a id="nestedobjatt--listener--http--redirects"></a>
### Nested Schema for `listener.http.redirects`

Read-Only:

- `http_to_https` (Boolean)



<a id="nestedobjatt--listener--stream"></a>
### Nested Schema for `listener.stream`

Read-Only:

- `handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--stream--handler))

<a id="nestedobjatt--listener--stream--handler"></a>
### Nested Schema for `listener.stream.handler`

Read-Only:

- `backend_group_id` (String)
- `idle_timeout` (String)



<a id="nestedobjatt--listener--tls"></a>
### Nested Schema for `listener.tls`

Read-Only:

- `default_handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--default_handler))
- `sni_handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--sni_handler))

<a id="nestedobjatt--listener--tls--default_handler"></a>
### Nested Schema for `listener.tls.default_handler`

Read-Only:

- `certificate_ids` (Set of String)
- `http_handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--default_handler--http_handler))
- `stream_handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--default_handler--stream_handler))

<a id="nestedobjatt--listener--tls--default_handler--http_handler"></a>
### Nested Schema for `listener.tls.default_handler.http_handler`

Read-Only:

- `allow_http10` (Boolean)
- `http2_options` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--default_handler--stream_handler--http2_options))
- `http_router_id` (String)
- `rewrite_request_id` (Boolean)

<a id="nestedobjatt--listener--tls--default_handler--stream_handler--http2_options"></a>
### Nested Schema for `listener.tls.default_handler.stream_handler.http2_options`

Read-Only:

- `max_concurrent_streams` (Number)



<a id="nestedobjatt--listener--tls--default_handler--stream_handler"></a>
### Nested Schema for `listener.tls.default_handler.stream_handler`

Read-Only:

- `backend_group_id` (String)
- `idle_timeout` (String)



<a id="nestedobjatt--listener--tls--sni_handler"></a>
### Nested Schema for `listener.tls.sni_handler`

Read-Only:

- `handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--sni_handler--handler))
- `name` (String)
- `server_names` (Set of String)

<a id="nestedobjatt--listener--tls--sni_handler--handler"></a>
### Nested Schema for `listener.tls.sni_handler.handler`

Read-Only:

- `certificate_ids` (Set of String)
- `http_handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--sni_handler--server_names--http_handler))
- `stream_handler` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--sni_handler--server_names--stream_handler))

<a id="nestedobjatt--listener--tls--sni_handler--server_names--http_handler"></a>
### Nested Schema for `listener.tls.sni_handler.server_names.http_handler`

Read-Only:

- `allow_http10` (Boolean)
- `http2_options` (List of Object) (see [below for nested schema](#nestedobjatt--listener--tls--sni_handler--server_names--http_handler--http2_options))
- `http_router_id` (String)
- `rewrite_request_id` (Boolean)

<a id="nestedobjatt--listener--tls--sni_handler--server_names--http_handler--http2_options"></a>
### Nested Schema for `listener.tls.sni_handler.server_names.http_handler.http2_options`

Read-Only:

- `max_concurrent_streams` (Number)



<a id="nestedobjatt--listener--tls--sni_handler--server_names--stream_handler"></a>
### Nested Schema for `listener.tls.sni_handler.server_names.stream_handler`

Read-Only:

- `backend_group_id` (String)
- `idle_timeout` (String)






<a id="nestedatt--log_options"></a>
### Nested Schema for `log_options`

Read-Only:

- `disable` (Boolean)
- `discard_rule` (List of Object) (see [below for nested schema](#nestedobjatt--log_options--discard_rule))
- `log_group_id` (String)

<a id="nestedobjatt--log_options--discard_rule"></a>
### Nested Schema for `log_options.discard_rule`

Read-Only:

- `discard_percent` (Number)
- `grpc_codes` (List of String)
- `http_code_intervals` (List of String)
- `http_codes` (List of Number)
