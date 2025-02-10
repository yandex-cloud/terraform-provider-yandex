//
// Get SmartCaptcha details by Id.
//
data "yandex_smartcaptcha_captcha" "by-id" {
  captcha_id = yandex_smartcaptcha_captcha.my-captcha.id
}
