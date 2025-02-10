//
// Create a new CDN Resource
//
resource "yandex_cdn_resource" "my_resource" {
  cname               = "cdn1.yandex-example.ru"
  active              = false
  origin_protocol     = "https"
  secondary_hostnames = ["cdn-example-1.yandex.ru", "cdn-example-2.yandex.ru"]
  origin_group_id     = yandex_cdn_origin_group.foo_cdn_group_by_id.id

  options {
    edge_cache_settings = 345600
    ignore_cookie       = true
    static_request_headers = {
      is-from-cdn = "yes"
    }
    static_response_headers = {
      is-cdn = "yes"
    }
  }
}
