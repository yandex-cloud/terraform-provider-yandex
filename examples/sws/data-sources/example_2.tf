data "yandex_sws_waf_profile" "by-name" {
  name = yandex_sws_waf_profile.my-profile.name
}
