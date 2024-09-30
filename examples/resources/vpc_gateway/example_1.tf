resource "yandex_vpc_gateway" "default" {
  name = "foobar"
  shared_egress_gateway {}
}
