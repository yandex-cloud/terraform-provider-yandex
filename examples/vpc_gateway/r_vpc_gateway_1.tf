//
// Create a new VPC NAT Gateway.
//
resource "yandex_vpc_gateway" "my_gw" {
  name = "foobar"
  shared_egress_gateway {}
}
