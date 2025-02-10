//
// Create a new VPC Subnet.
//
resource "yandex_vpc_subnet" "my_subnet" {
  v4_cidr_blocks = ["10.2.0.0/16"]
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.lab-net.id
}

resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}
