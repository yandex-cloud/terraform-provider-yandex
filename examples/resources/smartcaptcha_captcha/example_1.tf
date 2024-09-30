resource "yandex_smartcaptcha_captcha" "demo-captcha-simple" {
  deletion_protection = true
  name                = "demo-captcha-simple"
  complexity          = "HARD"
  pre_check_type      = "SLIDER"
  challenge_type      = "IMAGE_TEXT"

  allowed_sites = [
    "example.com",
    "example.ru"
  ]
}
