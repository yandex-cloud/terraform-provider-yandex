//
// Create a new DNS Firewall.
//
resource "yandex_dns_firewall" "fw1" {
  name        = "my-firewall"
  description = "desc"

  labels = {
    label1 = "label-1-value"
  }

  enabled         = true
  whitelist_fqdns = ["*.foo.bar."]
  blacklist_fqdns = ["bad.foo.bar."]

  resource_config = {
    resource_type  = "NETWORK"
    resource_ids   = [yandex_vpc_network.foo.id]
    lock_resources = true
  }

  deletion_protection = true
}

// Auxiliary resource for DNS Firewall
resource "yandex_vpc_network" "foo" {}
