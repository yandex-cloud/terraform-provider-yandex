data "yandex_sws_security_profile" "by-id" {
  security_profile_id = yandex_sws_security_profile.my-profile.id
}
