//
// Create a new VPC Private Endpoint.
//
resource "yandex_vpc_private_endpoint" "my_pe" {
  name        = "object-storage-private-endpoint"
  description = "description for private endpoint"

  labels = {
    my-label = "my-label-value"
  }

  network_id = yandex_vpc_network.lab-net.id

  object_storage {}

  dns_options {
    private_dns_records_enabled = true
  }

  endpoint_address {
    subnet_id = yandex_vpc_subnet.lab-subnet-a.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_subnet" "lab-subnet-a" {
  v4_cidr_blocks = ["10.2.0.0/16"]
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.lab-net.id
}
