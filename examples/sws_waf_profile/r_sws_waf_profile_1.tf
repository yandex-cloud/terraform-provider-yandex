//
// Create a new SWS WAF Profile (Minimal).
//
resource "yandex_sws_waf_profile" "minimal" {
  // NOTE: WAF profile must have at least one rule set, otherwise backend
  // rejects the request with `waf profile must have at least one rule`.
  // See the next example to see how to enable a default set of rules.
  name = "waf-profile-minimal"

  rule_set {
    action     = "DENY"
    is_enabled = true
    priority   = 1
    core_rule_set {
      inbound_anomaly_score = 2
      paranoia_level        = 1
      rule_set {
        name    = "OWASP Core Ruleset"
        version = "4.0.0"
        type    = "CORE"
      }
    }
  }
}
