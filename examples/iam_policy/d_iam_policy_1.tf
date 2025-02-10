//
// Get information about existing IAM Policy.
//
data "yandex_iam_policy" "admin" {
  binding {
    role = "admin"

    members = [
      "userAccount:user_id_1"
    ]
  }

  binding {
    role = "viewer"

    members = [
      "userAccount:user_id_2"
    ]
  }
}
