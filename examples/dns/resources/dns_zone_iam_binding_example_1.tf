resource "yandex_dns_zone" "zone1" {
  name = "my-private-zone"
  zone = "example.com."
}

resource "yandex_dns_zone_iam_binding" "viewer" {
  dns_zone_id = yandex_dns_zone.zone1.id
  role        = "dns.viewer"
  members     = ["userAccount:foo_user_id"]
}
