//
// Create a new MDB SQL Server Cluster.
//
resource "yandex_mdb_sqlserver_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "2016sp2std"

  resources {
    resource_preset_id = "s2.small"
    disk_type_id       = "network-ssd"
    disk_size          = 20
  }

  labels = { test_key : "test_value" }

  backup_window_start {
    hours   = 20
    minutes = 30
  }

  sqlserver_config = {
    fill_factor_percent           = 49
    optimize_for_ad_hoc_workloads = true
  }

  database {
    name = "db_name_a"
  }
  database {
    name = "db_name"
  }
  database {
    name = "db_name_b"
  }

  user {
    name     = "bob"
    password = "mysecurepassword"
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "db_name"
      roles         = ["DDLADMIN"]
    }
  }

  user {
    name     = "chuck"
    password = "mysecurepassword"

    permission {
      database_name = "db_name_a"
      roles         = ["OWNER"]
    }
    permission {
      database_name = "db_name"
      roles         = ["OWNER", "DDLADMIN"]
    }
    permission {
      database_name = "db_name_b"
      roles         = ["OWNER", "DDLADMIN"]
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }

  security_group_ids = [yandex_vpc_security_group.test-sg-x.id]
  host_group_ids     = ["host_group_1", "host_group_2"]
}

// Auxiliary resources
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
