//
// Get information about existing VPC Network.
//
data "yandex_vpc_network" "admin" {
  network_id = "my-network-id"
}
