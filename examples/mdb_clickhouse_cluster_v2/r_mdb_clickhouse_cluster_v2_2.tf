//
// Create a new MDB High Availability Clickhouse Cluster.
//
resource "yandex_mdb_clickhouse_cluster_v2" "my_cluster" {
  name        = "ha"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  clickhouse = {
    resources = {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  zookeeper = {
    resources = {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  hosts = {
    "clickhouse-host1" = {
      type      = "CLICKHOUSE"
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
    "clickhouse-host2" = {
      type      = "CLICKHOUSE"
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.bar.id
    }
    "keeper-host1" = {
      type      = "KEEPER"
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
    "keeper-host2" = {
      type      = "KEEPER"
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.bar.id
    }
    "keeper-host3" = {
      type      = "KEEPER"
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.baz.id
    }
  }

  cloud_storage = {
    enabled = false
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
