data "yandex_iam_role" "admin" {
  binding {
    role = "admin"

    members = [
      "userAccount:user_id_1"
    ]
  }
}
