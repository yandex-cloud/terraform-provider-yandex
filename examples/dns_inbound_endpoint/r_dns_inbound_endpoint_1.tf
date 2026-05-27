//
// Create a new DNS Inbound Endpoint.
//
resource "yandex_dns_inbound_endpoint" "endpoint1" {
  name        = "my-inbound-endpoint"
  description = "My DNS inbound endpoint"

  labels = {
    label1 = "label-1-value"
  }

  folder_id   = "my-folder-id"
  network_id  = yandex_vpc_network.foo.id
  address_id  = yandex_vpc_address.addr1.id

  deletion_protection = true
}

// Auxiliary resources for DNS Inbound Endpoint
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "my-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "my-addr"
  description = "internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}
