//
// Create a new DNS Firewall and new IAM Binding for it.
//
resource "yandex_dns_firewall" "fw1" {
  name = "my-firewall"
}

resource "yandex_dns_firewall_iam_binding" "fw-editor" {
  dns_firewall_id = yandex_dns_firewall.fw1.id
  role        = "dns.firewallEditor"
  members     = ["userAccount:foo_user_id"]
}
