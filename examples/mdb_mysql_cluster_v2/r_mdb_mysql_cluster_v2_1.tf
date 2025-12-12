//
// Create a new MDB MySQL Cluster (v2).
//

resource "yandex_mdb_mysql_cluster_v2" "cluster" {
  name        = "mysql-cluster"
  description = "MySQL Test Cluster"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id
  environment = "PRODUCTION"

  labels = {
    "key1" = "value1"
    "key2" = "value2"
    "key3" = "value3"
  }

  hosts = {
    "host" = {
      zone             = "ru-central1-a"
      subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
      assign_public_ip = false
    }
  }

  version = "8.0"
  resources {
    resource_preset_id = "b1.medium"
    disk_type_id       = "network-ssd"
    disk_size          = 10
  }

  performance_diagnostics = {
    sessions_sampling_interval   = 60
    statements_sampling_interval = 600
  }

  access = {
    web_sql       = true
    data_transfer = true
    data_lens     = true
    yandex_query  = true
  }

  maintenance_window = {
    type = "WEEKLY"
    day  = "MON"
    hour = 3
  }

  backup_window_start = {
    hours   = 5
    minutes = 5
  }

  backup_retain_period_days = 8
  deletion_protection       = true
}

// Auxiliary resources
resource "yandex_vpc_network" "test-net" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_security_group" "test-sgroup" {
  description = "Test security group"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id
}
