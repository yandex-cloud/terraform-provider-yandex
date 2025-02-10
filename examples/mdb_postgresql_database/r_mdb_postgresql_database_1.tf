//
// Create a new MDB PostgreSQL Database.
//
resource "yandex_mdb_postgresql_database" "my_db" {
  cluster_id = yandex_mdb_postgresql_cluster.my_cluster.id
  name       = "testdb"
  owner      = yandex_mdb_postgresql_user.my_user.name
  lc_collate = "en_US.UTF-8"
  lc_type    = "en_US.UTF-8"
  extension {
    name = "uuid-ossp"
  }
  extension {
    name = "xml2"
  }
}

resource "yandex_mdb_postgresql_user" "my_user" {
  cluster_id = yandex_mdb_postgresql_cluster.my_cluster.id
  name       = "alice"
  password   = "password"
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
