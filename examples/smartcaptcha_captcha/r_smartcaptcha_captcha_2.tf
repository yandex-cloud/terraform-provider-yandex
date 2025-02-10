//
// Advanced SmartCaptcha example.
//
resource "yandex_smartcaptcha_captcha" "demo-captcha-advanced" {
  deletion_protection = true
  name                = "demo-captcha-advanced"
  complexity          = "HARD"
  pre_check_type      = "SLIDER"
  challenge_type      = "IMAGE_TEXT"

  allowed_sites = [
    "example.com",
    "example.ru"
  ]

  override_variant {
    uuid        = "xxx"
    description = "override variant 1"

    complexity     = "EASY"
    pre_check_type = "CHECKBOX"
    challenge_type = "SILHOUETTES"
  }

  override_variant {
    uuid        = "yyy"
    description = "override variant 2"

    complexity     = "HARD"
    pre_check_type = "CHECKBOX"
    challenge_type = "KALEIDOSCOPE"
  }

  security_rule {
    name                  = "rule1"
    priority              = 11
    description           = "My first security rule. This rule it's just example to show possibilities of configuration."
    override_variant_uuid = "xxx"

    condition {
      host {
        hosts {
          exact_match = "example.com"
        }
        hosts {
          exact_match = "example.net"
        }
      }

      uri {
        path {
          prefix_match = "/form"
        }
        queries {
          key = "firstname"
          value {
            pire_regex_match = ".*ivan.*"
          }
        }
        queries {
          key = "lastname"
          value {
            pire_regex_match = ".*petr.*"
          }
        }
      }

      headers {
        name = "User-Agent"
        value {
          pire_regex_match = ".*curl.*"
        }
      }
      headers {
        name = "Referer"
        value {
          pire_regex_not_match = ".*bot.*"
        }
      }

      source_ip {
        ip_ranges_match {
          ip_ranges = ["1.2.33.44", "2.3.4.56"]
        }
        ip_ranges_not_match {
          ip_ranges = ["8.8.0.0/16", "10::1234:1abc:1/64"]
        }
        geo_ip_match {
          locations = ["ru", "es"]
        }
        geo_ip_not_match {
          locations = ["us", "fm", "gb"]
        }
      }
    }
  }

  security_rule {
    name                  = "rule2"
    priority              = 555
    description           = "Second rule"
    override_variant_uuid = "yyy"

    condition {
      uri {
        path {
          prefix_match = "/form"
        }
      }
    }
  }

  security_rule {
    name                  = "rule3"
    priority              = 99999
    description           = "Empty condition rule"
    override_variant_uuid = "yyy"
  }
}
