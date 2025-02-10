//
// Get information about existing VPC Security Group.
//
data "yandex_vpc_security_group" "group1" {
  security_group_id = "my-id"
}

data "yandex_vpc_security_group" "group1" {
  name = "my-group1"
}
