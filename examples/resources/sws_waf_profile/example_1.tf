resource "yandex_sws_waf_profile" "empty" {
  // NOTE: this WAF profile do not contains any rules enabled.
  // See the next example to see how to enable default set of rules. 
  name = "waf-profile-dummy"

  core_rule_set {
    inbound_anomaly_score = 2
    paranoia_level        = 1
    rule_set {
      name    = "OWASP Core Ruleset"
      version = "4.0.0"
    }
  }
}
