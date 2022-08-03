---
layout: "yandex"
page_title: "Yandex: yandex_cdn_resource"
sidebar_current: "docs-yandex-cnd-resource"
description: |-
 Allows management of a Yandex.Cloud CDN Resource.
---

# yandex\_cdn\_resource

Allows management of [Yandex.Cloud CDN Resource](https://cloud.yandex.ru/docs/cdn/concepts/resource).

> **_NOTE:_**  CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: ```yc cdn provider activate --folder-id <folder-id> --type gcore```

## Example Usage

```hcl
resource "yandex_cdn_resource" "my_resource" {
	cname = "cdn1.yandex-example.ru"

	active = false

	origin_protocol = "https"

	secondary_hostnames = ["cdn-example-1.yandex.ru", "cdn-example-2.yandex.ru"]

	origin_group_id = yandex_cdn_origin_group.foo_cdn_group_by_id.id

    options {
        edge_cache_settings = 345600
        ignore_cookie = true
    }
}
```

## Argument Reference

The following arguments are supported:

* `cname` (Required) - CDN endpoint CNAME, must be unique among resources.

* `active` (Optional) - Flag to create Resource either in active or disabled state. True - the content from CDN is available to clients.

* `options` (Optional) - CDN Resource settings and options to tune CDN edge behavior.

* `secondary_hostnames` (Optional) - list of secondary hostname strings.

* `ssl_certificate` (Optional) - SSL certificate of CDN resource.
---

Resource block supports following options:

* `disable_cache` - setup a cache status.

* `edge_cache_settings` - content will be cached according to origin cache settings. The value applies for a response with codes 200, 201, 204, 206, 301, 302, 303, 304, 307, 308 if an origin server does not have caching HTTP headers. Responses with other codes will not be cached.

* `browser_cache_settings` - set up a cache period for the end-users browser. Content will be cached due to origin settings. If there are no cache settings on your origin, the content will not be cached. The list of HTTP response codes that can be cached in browsers: 200, 201, 204, 206, 301, 302, 303, 304, 307, 308. Other response codes will not be cached. The default value is 4 days.

* `cache_http_headers` - list HTTP headers that must be included in responses to clients.

* `ignore_query_params` - files with different query parameters are cached as objects with the same key regardless of the parameter value. selected by default.

* `query_params_whitelist` - files with the specified query parameters are cached as objects with different keys, files with other parameters are cached as objects with the same key.

* `query_params_blacklist` - files with the specified query parameters are cached as objects with the same key, files with other parameters are cached as objects with different keys.

* `slice` - files larger than 10 MB will be requested and cached in parts (no larger than 10 MB each part). It reduces time to first byte. The origin must support HTTP Range requests.

* `fetched_compressed` - option helps you to reduce the bandwidth between origin and CDN servers. Also, content delivery speed becomes higher because of reducing the time for compressing files in a CDN.

* `gzip_on` - GZip compression at CDN servers reduces file size by 70% and can be as high as 90%.

* `redirect_http_to_https` - set up a redirect from HTTP to HTTPS.

* `redirect_https_to_http` - set up a redirect from HTTPS to HTTP.

* `custom_host_header` - custom value for the Host header. Your server must be able to process requests with the chosen header.

* `forward_host_header` - choose the Forward Host header option if is important to send in the request to the Origin the same Host header as was sent in the request to CDN server.

* `cors` - parameter that lets browsers get access to selected resources from a domain different to a domain from which the request is received.

* `stale` -  list of errors which instruct CDN servers to serve stale content to clients.

* `allowed_http_methods` - HTTP methods for your CDN content. By default the following methods are allowed: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS. In case some methods are not allowed to the user, they will get the 405 (Method Not Allowed) response. If the method is not supported, the user gets the 501 (Not Implemented) response.

* `proxy_cache_methods_set` - allows caching for GET, HEAD and POST requests.

* `disable_proxy_force_ranges` - disabling proxy force ranges.

* `static_request_headers` - set up custom headers that CDN servers send in requests to origins.

* `custom_server_name` - wildcard additional CNAME. If a resource has a wildcard additional CNAME, you can use your own certificate for content delivery via HTTPS. Read-only.

* `ignore_cookie` - set for ignoring cookie.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the IoT Core Device

## Timeouts

This resource provides the following configuration options for
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

A origin group can be imported using any of these accepted formats:

```
$ terraform import yandex_cdn_resource.default origin_group_id
```
