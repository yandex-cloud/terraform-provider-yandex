---
subcategory: "Cloud Content Delivery Network (CDN)"
page_title: "Yandex: yandex_cdn_resource"
description: |-
  Get information about a Yandex CDN Resource.
---

# yandex_cdn_resource (Data Source)

Get information about a Yandex CDN Resource. For more information, see [the official documentation](https://yandex.cloud/docs/cdn/concepts/resource).

~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: `yc cdn provider activate --folder-id <folder-id> --type gcore`

## Example usage

```terraform
data "yandex_cdn_resource" "my_resource" {
  resource_id = "some resource id"
}

output "resource_cname" {
  value = data.yandex_cdn_resource.my_resource.cname
}
```

## Argument Reference

The following arguments are supported:

* `cname` (Required) - CDN endpoint CNAME, must be unique among resources.

* `folder_id` (Optional) - Folder that the resource belongs to. If value is omitted, the default provider folder is used.

* `labels` - (Optional) Labels to assign to this CDN Resource. A list of key/value pairs.

* `active` (Optional) - Flag to create Resource either in active or disabled state. True - the content from CDN is available to clients.

* `options` (Optional) - CDN Resource settings and options to tune CDN edge behavior.

* `secondary_hostnames` (Optional) - list of secondary hostname strings.

* `ssl_certificate` (Optional) - SSL certificate of CDN resource.

* `provider_cname` (Optional) - provider CNAME of CDN resource, computed value for read and update operations.

---

Resource options block supports following options:

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

* `redirect_http_to_https` - set up a redirect from HTTPS to HTTP.

* `redirect_https_to_http` - set up a redirect from HTTP to HTTPS.

* `custom_host_header` - custom value for the Host header. Your server must be able to process requests with the chosen header.

* `forward_host_header` - choose the Forward Host header option if is important to send in the request to the Origin the same Host header as was sent in the request to CDN server.

* `cors` - parameter that lets browsers get access to selected resources from a domain different to a domain from which the request is received.

* `stale` - list of errors which instruct CDN servers to serve stale content to clients.

* `allowed_http_methods` - HTTP methods for your CDN content. By default the following methods are allowed: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS. In case some methods are not allowed to the user, they will get the 405 (Method Not Allowed) response. If the method is not supported, the user gets the 501 (Not Implemented) response.

* `proxy_cache_methods_set` - allows caching for GET, HEAD and POST requests.

* `disable_proxy_force_ranges` - disabling proxy force ranges.

* `static_request_headers` - set up custom headers that CDN servers send in requests to origins.

* `custom_server_name` - wildcard additional CNAME. If a resource has a wildcard additional CNAME, you can use your own certificate for content delivery via HTTPS. Read-only.

* `ignore_cookie` - set for ignoring cookie.

* `secure_key` - set secure key for url encoding to protect contect and limit access by IP addresses and time limits.

* `enable_ip_url_signing` - enable access limiting by IP addresses, option available only with setting secure_key.

* `ip_address_acl.excepted_values` - the list of specified IP addresses to be allowed or denied depending on acl policy type.

* `ip_address_acl.policy_type` - the policy type for ip_address_acl option, one of "allow" or "deny" values.
