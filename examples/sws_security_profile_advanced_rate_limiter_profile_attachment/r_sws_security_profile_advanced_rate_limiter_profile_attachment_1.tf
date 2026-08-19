resource "yandex_sws_advanced_rate_limiter_profile" "arl" {
  name = "example-arl-profile"

  advanced_rate_limiter_rule {
    name     = "example-rule"
    priority = 1
    dry_run  = true

    static_quota {
      action = "DENY"
      limit  = 1000
      period = 60
    }
  }
}

resource "yandex_sws_security_profile" "security_profile" {
  name                     = "example-security-profile"
  default_action           = "ALLOW"
  disallow_data_processing = false
}

resource "yandex_sws_security_profile_advanced_rate_limiter_profile_attachment" "attachment" {
  security_profile_id              = yandex_sws_security_profile.security_profile.id
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.arl.id
}
