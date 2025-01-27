resource "yandex_sws_advanced_rate_limiter_profile" "demo-profile" {
  name = "demo-profile"

  advanced_rate_limiter_rule {
    name        = "rule1"
    priority    = 10
    description = "First test rule"
    dry_run     = true

    static_quota {
      action = "DENY"
      limit  = 10000000
      period = 1
      condition {
        request_uri {
          path {
            exact_match = "/api"
          }
        }
      }
    }
  }
}
