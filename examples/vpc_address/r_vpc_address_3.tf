//
// Create a new VPC internal IPv4 Address.
// The address can be used in compute_instance, vpc_private_endpoint or lb_network_load_balancer resources.
//
resource "yandex_vpc_address" "internal_addr" {
  name = "internalAddress"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

// Auxiliary resources for VPC Address
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
