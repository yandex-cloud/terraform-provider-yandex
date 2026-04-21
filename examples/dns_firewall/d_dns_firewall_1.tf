//
// Get information about existing DNS Firewall.
//
data "yandex_dns_firewall" "foo" {
  dns_firewall_id = yandex_dns_firewall.fw1.id
}

output "whitelist_fqdns" {
  value = data.yandex_dns_firewall.foo.whitelist_fqdns
}
