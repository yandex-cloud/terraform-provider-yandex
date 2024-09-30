data "yandex_dns_zone" "foo" {
  dns_zone_id = yandex_dns_zone.zone1.id
}

output "zone" {
  value = data.yandex_dns_zone.foo.zone
}
