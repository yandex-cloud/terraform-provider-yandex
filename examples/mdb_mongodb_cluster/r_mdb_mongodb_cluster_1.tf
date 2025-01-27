resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_mdb_mongodb_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  cluster_config {
    version = "4.2"
  }

  labels = {
    test_key = "test_value"
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  resources_mongod {
    resource_preset_id = "s2.small"
    disk_size          = 16
    disk_type_id       = "network-hdd"
  }

  resources_mongos {
    resource_preset_id = "s2.small"
    disk_size          = 14
    disk_type_id       = "network-hdd"
  }

  resources_mongocfg {
    resource_preset_id = "s2.small"
    disk_size          = 14
    disk_type_id       = "network-hdd"
  }

  host {
    zone_id   = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }

  maintenance_window {
    type = "ANYTIME"
  }
}
