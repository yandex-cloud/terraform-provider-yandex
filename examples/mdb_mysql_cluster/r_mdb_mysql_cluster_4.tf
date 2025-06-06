//
// Create a new MDB MySQL Cluster with different backup priorities.
//
resource "yandex_mdb_mysql_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }

  host {
    zone      = "ru-central1-b"
    name      = "na-1"
    subnet_id = yandex_vpc_subnet.foo.id
  }
  host {
    zone            = "ru-central1-d"
    name            = "nb-1"
    backup_priority = 5
    subnet_id       = yandex_vpc_subnet.bar.id
  }
  host {
    zone            = "ru-central1-d"
    name            = "nb-2"
    backup_priority = 10
    subnet_id       = yandex_vpc_subnet.bar.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}
