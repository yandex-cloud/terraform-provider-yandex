resource "yandex_mdb_greenplum_cluster" "foo" {
  name               = "test"
  description        = "test greenplum cluster"
  environment        = "PRESTABLE"
  network_id         = yandex_vpc_network.foo.id
  zone_id            = "ru-central1-a"
  subnet_id          = yandex_vpc_subnet.foo.id
  assign_public_ip   = true
  version            = "6.22"
  master_host_count  = 2
  segment_host_count = 5
  segment_in_host    = 1
  master_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }
  segment_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }

  access {
    web_sql = true
  }

  greenplum_config = {
    max_connections         = 395
    gp_workfile_compression = "false"
  }

  user_name     = "admin_user"
  user_password = "your_super_secret_password"

  security_group_ids = [yandex_vpc_security_group.test-sg-x.id]
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}

resource "yandex_vpc_security_group" "test-sg-x" {
  network_id = yandex_vpc_network.foo.id
  ingress {
    protocol       = "ANY"
    description    = "Allow incoming traffic from members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol       = "ANY"
    description    = "Allow outgoing traffic to members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
