//
// Get information about existing SWS WAF Profile.
//
data "yandex_sws_waf_profile" "by-id" {
  waf_profile_id = yandex_sws_waf_profile.my-profile.id
}

data "yandex_sws_waf_profile" "by-name" {
  name = yandex_sws_waf_profile.my-profile.name
}
