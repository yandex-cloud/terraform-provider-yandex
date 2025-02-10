//
// Create a new MDB MongoDB User.
//
resource "yandex_mdb_mongodb_user" "my_user" {
  cluster_id = yandex_mdb_mongodb_cluster.my_cluster.id
  name       = "alice"
  password   = "password"
}

resource "yandex_mdb_mongodb_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  cluster_config {
    version = "6.0"
  }

  host {
    zone_id   = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo.id
  }
  resources_mongod {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
