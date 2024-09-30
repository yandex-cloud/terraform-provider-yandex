data "yandex_sws_waf_profile" "by-id" {
  waf_profile_id = yandex_sws_waf_profile.my-profile.id
}
