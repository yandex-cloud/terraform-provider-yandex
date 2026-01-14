//
// Create a new MDB Greenplum Cluster.
//
resource "yandex_mdb_greenplum_cluster_v2" "my_cluster" {
  depends_on = [yandex_vpc_subnet.foo]

  name        = "test"
  description = "test greenplum cluster"
  environment = "PRESTABLE"

  segment_host_count = 2
  segment_in_host   = 1

  user_name = "test-user"
  user_password = "test-user-password"
  network_id = yandex_vpc_network.foo.id

  cluster_config = {
    assign_public_ip = true
    backup_window_start = {
      hours   = 1
      minutes = 30
    }
  }

  config = {
    zone_id = "ru-central1-a"
  }

  master_config = {
    resources = {
      resource_preset_id = "s2.small"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  segment_config = {
    resources = {
      resource_preset_id = "s2.small"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  cloud_storage = {
    enable = true
  }
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
