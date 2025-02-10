//
// Create a new CM Certificate IAM Binding.
//
resource "yandex_cm_certificate" "your-certificate" {
  name    = "certificate-name"
  domains = ["example.com"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}

resource "yandex_cm_certificate_iam_binding" "viewer_binding" {
  certificate_id = yandex_cm_certificate.your-certificate.id
  role           = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
