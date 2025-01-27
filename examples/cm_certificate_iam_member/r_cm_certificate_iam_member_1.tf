resource "yandex_cm_certificate" "your-certificate" {
  name = "certificate-name"
  domains = ["example.com"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}

resource "yandex_cm_certificate_iam_member" "viewer" {
  certificate_id = yandex_cm_certificate.your-certificate.id
  role      = "viewer"

  member = "userAccount:foo_user_id"
}
