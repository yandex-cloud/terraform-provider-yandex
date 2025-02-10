//
// Create a new VPC Security Group Rule.
//
resource "yandex_vpc_security_group" "group1" {
  name        = "My security group"
  description = "description for my security group"
  network_id  = yandex_vpc_network.lab-net.id

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_vpc_security_group_rule" "rule1" {
  security_group_binding = yandex_vpc_security_group.group1.id
  direction              = "ingress"
  description            = "rule1 description"
  v4_cidr_blocks         = ["10.0.1.0/24", "10.0.2.0/24"]
  port                   = 8080
  protocol               = "TCP"
}

resource "yandex_vpc_security_group_rule" "rule2" {
  security_group_binding = yandex_vpc_security_group.group1.id
  direction              = "egress"
  description            = "rule2 description"
  v4_cidr_blocks         = ["10.0.1.0/24"]
  from_port              = 8090
  to_port                = 8099
  protocol               = "UDP"
}

// Auxiliary resources
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}
