//
// Attach a VPC network and an existing Cloud Interconnect private connection.
//
variable "cic_private_connection_id" {
  type        = string
  description = "ID of an existing Cloud Interconnect private connection to attach."
}

resource "yandex_cloudrouter_routing_instance" "with_private_connection" {
  name        = "my-interconnect-routing-instance"
  description = "Routing between a VPC network and a private connection"

  labels = {
    environment = "production"
  }

  // Each configured collection replaces its full set of entries.
  // Omit a collection to keep its current entries. Set it to [] to remove all.
  vpc_info = [{
    vpc_network_id = yandex_vpc_network.interconnect.id
    az_infos = [{
      manual_info = {
        az_id    = yandex_vpc_subnet.interconnect.zone
        prefixes = yandex_vpc_subnet.interconnect.v4_cidr_blocks
      }
    }]
  }]

  cic_private_connection_info = [{
    cic_private_connection_id = var.cic_private_connection_id
  }]

  // Disable deletion protection before destroying the routing instance.
  deletion_protection = true
}

// Auxiliary resource for the routing instance.
resource "yandex_vpc_network" "interconnect" {
  name = "my-interconnect-network"
}

resource "yandex_vpc_subnet" "interconnect" {
  name           = "my-interconnect-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.interconnect.id
  v4_cidr_blocks = ["10.10.0.0/24"]
}
