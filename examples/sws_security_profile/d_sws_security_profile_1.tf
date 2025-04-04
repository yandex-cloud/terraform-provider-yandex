//
// Get information about existing SWS Security Profile.
//
data "yandex_sws_security_profile" "by-id" {
  security_profile_id = yandex_sws_security_profile.my-profile.id
}

data "yandex_sws_security_profile" "by-name" {
  name = yandex_sws_security_profile.my-profile.name
}
