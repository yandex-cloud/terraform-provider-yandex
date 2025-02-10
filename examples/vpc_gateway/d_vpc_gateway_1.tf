//
// Get information about existing VPC NAT Gateway.
//
data "yandex_vpc_gateway" "default" {
  gateway_id = "my-gateway-id"
}
