//
// Create a new CDN Rule
//
resource "yandex_cdn_rule" "my_rule" {
  resource_id  = yandex_cdn_resource.my_resource.id
  name         = "images-rule"
  rule_pattern = "/images/.*"
  weight       = 100

  options {
    edge_cache_settings {
      enabled    = true
      cache_time = {
        "*" = 86400
      }
    }

    browser_cache_settings {
      enabled    = true
      cache_time = 3600
    }

    gzip_on = true
  }
}
