//
// Example of Certificate Validation. 
// Use "data.yandex_cm_certificate.example.id" to get validated certificate.
//
resource "yandex_cm_certificate" "example" {
  name    = "example"
  domains = ["example.com", "*.example.com"]

  managed {
    challenge_type  = "DNS_CNAME"
    challenge_count = 1 # "example.com" and "*.example.com" has the same challenge
  }
}

resource "yandex_dns_recordset" "example" {
  count   = yandex_cm_certificate.example.managed[0].challenge_count
  zone_id = "example-zone-id"
  name    = yandex_cm_certificate.example.challenges[count.index].dns_name
  type    = yandex_cm_certificate.example.challenges[count.index].dns_type
  data    = [yandex_cm_certificate.example.challenges[count.index].dns_value]
  ttl     = 60
}

data "yandex_cm_certificate" "example" {
  depends_on      = [yandex_dns_recordset.example]
  certificate_id  = yandex_cm_certificate.example.id
  wait_validation = true
}
