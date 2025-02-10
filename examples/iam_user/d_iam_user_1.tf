//
// Get information about existing IAM User.
//
data "yandex_iam_user" "admin" {
  login = "my-yandex-login"
}
