//
// Create a new MDB PostgreSQL database User.
//
resource "yandex_mdb_postgresql_user" "my_user" {
  cluster_id = yandex_mdb_postgresql_cluster.my_cluster.id
  name       = "alice"
  password   = "password"
  conn_limit = 50
  settings = {
    default_transaction_isolation = "read committed"
    log_min_duration_statement    = 5000
  }
}

resource "yandex_mdb_postgresql_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 15
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
