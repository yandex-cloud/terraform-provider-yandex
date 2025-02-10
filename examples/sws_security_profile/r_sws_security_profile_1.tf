//
// Create a new SWS Security Profile (Simple).
//
resource "yandex_sws_security_profile" "demo-profile-simple" {
  name           = "demo-profile-simple"
  default_action = "ALLOW"

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "API"
    }
  }
}
