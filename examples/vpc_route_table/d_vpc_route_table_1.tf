//
// Get information about existing VPC Route Table.
//
data "yandex_vpc_route_table" "my_rt" {
  route_table_id = "my-rt-id"
}
