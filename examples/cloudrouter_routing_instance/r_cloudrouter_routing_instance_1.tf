//
// Create a routing instance and attach a VPC network.
//
resource "yandex_cloudrouter_routing_instance" "example" {
  name = "my-routing-instance"

  vpc_info = [{
    vpc_network_id = yandex_vpc_network.example.id
  }]
}

// Auxiliary resource for the routing instance.
resource "yandex_vpc_network" "example" {
  name = "my-network"
}
