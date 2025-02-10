//
// Get information about existing VPC IPv4 Address.
//
data "yandex_vpc_address" "addr" {
  address_id = "my-address-id"
}
