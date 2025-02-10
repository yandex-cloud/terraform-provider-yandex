//
// Create a new CM Certificate IAM Member.
//
resource "yandex_cm_certificate" "your-certificate" {
  name    = "certificate-name"
  domains = ["example.com"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}

resource "yandex_cm_certificate_iam_member" "viewer_member" {
  certificate_id = yandex_cm_certificate.your-certificate.id
  role           = "viewer"

  member = "userAccount:foo_user_id"
}
