resource "yandex_sws_waf_profile" "waf" {
  name = "example-waf-profile"

  core_rule_set {
    inbound_anomaly_score = 2
    paranoia_level        = 4
    rule_set {
      name    = "OWASP Core Ruleset"
      version = "4.0.0"
      type    = "CORE"
    }
  }

  analyze_request_body {
    is_enabled        = true
    size_limit        = 8
    size_limit_action = "IGNORE"
  }

  rule_set {
    action     = "DENY"
    is_enabled = true
    priority   = 1
    core_rule_set {
      inbound_anomaly_score = 2
      paranoia_level        = 4
      rule_set {
        name    = "OWASP Core Ruleset"
        version = "4.0.0"
        type    = "CORE"
        id      = "OWASP_CRS_4_0_0"
      }
    }
  }
}

resource "yandex_sws_security_profile" "security_profile" {
  name                     = "example-security-profile"
  default_action           = "ALLOW"
  disallow_data_processing = false

  security_rule {
    name     = "waf"
    priority = 1

    waf {
      mode           = "FULL"
      waf_profile_id = yandex_sws_waf_profile.waf.id
    }
  }
}

resource "yandex_sws_security_profile_waf_profile_attachment" "attachment" {
  security_profile_id = yandex_sws_security_profile.security_profile.id
  waf_profile_id      = yandex_sws_waf_profile.waf.id
  security_rule_name  = "waf"
}
