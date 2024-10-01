resource "yandex_cm_certificate" "example" {
  name    = "example"
  domains = ["example.com"]

  managed {
    challenge_type = "DNS_CNAME"
  }
}
