//
// Get SmartCaptcha details by Name
//
data "yandex_smartcaptcha_captcha" "by-name" {
  name = yandex_smartcaptcha_captcha.my-captcha.name
}
