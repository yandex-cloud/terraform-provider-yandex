resource "yandex_cm_certificate" "example" {
  name    = "example"
  domains = ["one.example.com", "two.example.com"]

  managed {
    challenge_type  = "DNS_CNAME"
    challenge_count = 2 # for each domain
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
