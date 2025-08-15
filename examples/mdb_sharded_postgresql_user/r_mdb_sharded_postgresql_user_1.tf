//
// Create a new MDB Sharded PostgreSQL database User.
//
resource "yandex_mdb_sharded_postgresql_user" "my_user" {
  cluster_id = yandex_mdb_sharded_postgresql_cluster.my_user.id
  name       = "alice"
  password   = "password"
  settings = {
    connection_limit = 300
    connection_retries = 5
  }
}

resource "yandex_mdb_sharded_postgresql_cluster" "default" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config = {
    sharded_postgresql_config = {
      common = {
        console_password = "P@ssw0rd"
      }
      router = {
        resources = {
          resource_preset_id = "s2.micro"
          disk_type_id       = "network-ssd"
          disk_size          = 32
        }
      }
    }
  }

  hosts = {
    "router1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
      assign_public_ip = false
      type = "ROUTER"
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
