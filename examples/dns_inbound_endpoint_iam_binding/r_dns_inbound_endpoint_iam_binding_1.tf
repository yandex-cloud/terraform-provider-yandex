//
// Create a new DNS Inbound Endpoint and new IAM Binding for it.
//
resource "yandex_vpc_network" "network1" {
  name = "my-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "my-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
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

resource "yandex_dns_inbound_endpoint" "endpoint1" {
  name       = "my-inbound-endpoint"
  network_id = yandex_vpc_network.network1.id
  address_id = yandex_vpc_address.addr1.id
}

resource "yandex_dns_inbound_endpoint_iam_binding" "endpoint-editor" {
  dns_inbound_endpoint_id = yandex_dns_inbound_endpoint.endpoint1.id
  role                    = "dns.endpointEditor"
  members                 = ["userAccount:foo_user_id"]
}
