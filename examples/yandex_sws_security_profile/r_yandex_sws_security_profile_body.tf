resource "yandex_sws_security_profile" "body" {
  name           = "body-filter"
  default_action = "ALLOW"

  analyze_request_body {
    size_limit        = 8
    size_limit_action = "IGNORE"
  }

  security_rule {
    name     = "block-php2shell"
    priority = 352

    rule_condition {
      action = "DENY"

      condition {
        http_method {
          http_method_matcher {
            exact_match = "POST"
          }
        }

        body {
          body_matcher {
            pire_regex_match = ".*php2shell.*"
          }
        }
      }
    }
  }
}
