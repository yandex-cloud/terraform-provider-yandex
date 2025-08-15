//
// Create a new MDB Sharded PostgreSQL Cluster.
//
resource "yandex_mdb_sharded_postgresql_cluster" "default" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config = {
    backup_retain_period_days = 10
    sharded_postgresql_config = {
      common = {
        console_password = "P@ssw0rd"
        log_level = "INFO"
      }
      router = {
        resources = {
          resource_preset_id = "s2.micro"
          disk_type_id       = "network-ssd"
          disk_size          = 32
        }
        config = {
          show_notice_messages = false
          prefer_same_availability_zone = true
        }
      }
      coordinator = {
        resources = {
          resource_preset_id = "s2.micro"
          disk_type_id       = "network-ssd"
          disk_size          = 32
        }
      }
      balancer = {}
    }
  }

  hosts = {
    "router1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
      assign_public_ip = false
      type = "ROUTER"
    }
    "router2" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.foo.id
      assign_public_ip = false
      type = "ROUTER"
    }
    "coordinator" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.foo.id
      assign_public_ip = false
      type = "COORDINATOR"
    }
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
