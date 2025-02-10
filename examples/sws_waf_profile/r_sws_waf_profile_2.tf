//
// Create a new SWS WAF Profile (Default).
//
locals {
  waf_paranoia_level = 1
}

data "yandex_sws_waf_rule_set_descriptor" "owasp4" {
  name    = "OWASP Core Ruleset"
  version = "4.0.0"
}

resource "yandex_sws_waf_profile" "default" {
  name = "waf-profile-default"

  core_rule_set {
    inbound_anomaly_score = 2
    paranoia_level        = local.waf_paranoia_level
    rule_set {
      name    = "OWASP Core Ruleset"
      version = "4.0.0"
    }
  }

  dynamic "rule" {
    for_each = [
      for rule in data.yandex_sws_waf_rule_set_descriptor.owasp4.rules : rule
      if rule.paranoia_level >= local.waf_paranoia_level
    ]
    content {
      rule_id     = rule.value.id
      is_enabled  = true
      is_blocking = false
    }
  }

  analyze_request_body {
    is_enabled        = true
    size_limit        = 8
    size_limit_action = "IGNORE"
  }
}
