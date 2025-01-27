data "yandex_sws_advanced_rate_limiter_profile" "by-id" {
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.my-profile.id
}
