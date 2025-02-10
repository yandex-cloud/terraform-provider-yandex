//
// Create a new VPC Security Group.
//
resource "yandex_vpc_security_group" "sg1" {
  name        = "My security group"
  description = "description for my security group"
  network_id  = yandex_vpc_network.lab-net.id

  labels = {
    my-label = "my-label-value"
  }

  ingress {
    protocol       = "TCP"
    description    = "rule1 description"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    protocol       = "ANY"
    description    = "rule2 description"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }

  egress {
    protocol       = "UDP"
    description    = "rule3 description"
    v4_cidr_blocks = ["10.0.1.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

