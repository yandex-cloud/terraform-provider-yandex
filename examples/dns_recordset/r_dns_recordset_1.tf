//
// Create a new DNS Zone & two DNS Records on it.
//
resource "yandex_dns_zone" "zone1" {
  name        = "my_private_zone"
  description = "desc"

  labels = {
    label1 = "label-1-value"
  }

  zone             = "example.com."
  public           = false
  private_networks = [yandex_vpc_network.foo.id]
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv.example.com."
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.1"]
}

resource "yandex_dns_recordset" "rs2" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv2"
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.2"]
}

// Auxiliary resource for DNS Zone
resource "yandex_vpc_network" "foo" {}
