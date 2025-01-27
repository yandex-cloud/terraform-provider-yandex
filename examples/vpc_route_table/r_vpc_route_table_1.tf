resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_gateway" "egress-gateway" {
  name = "egress-gateway"
  shared_egress_gateway {}
}

resource "yandex_vpc_route_table" "lab-rt-a" {
  network_id = yandex_vpc_network.lab-net.id

  static_route {
    destination_prefix = "10.2.0.0/16"
    next_hop_address   = "172.16.10.10"
  }

  static_route {
    destination_prefix = "0.0.0.0/0"
    gateway_id         = yandex_vpc_gateway.egress-gateway.id
  }
}
