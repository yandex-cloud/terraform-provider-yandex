//
// Get information about existing VPC Subnet.
//
data "yandex_vpc_subnet" "admin" {
  subnet_id = "my-subnet-id"
}
