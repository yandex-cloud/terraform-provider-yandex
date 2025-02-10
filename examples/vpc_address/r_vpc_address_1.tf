//
// Create a new VPC regular IPv4 Address.
//
resource "yandex_vpc_address" "addr" {
  name = "exampleAddress"

  external_ipv4_address {
    zone_id = "ru-central1-a"
  }
}
