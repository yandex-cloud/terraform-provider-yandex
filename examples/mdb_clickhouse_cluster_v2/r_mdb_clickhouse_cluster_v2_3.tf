//
// Create a new MDB Sharded Clickhouse Cluster.
//
resource "yandex_mdb_clickhouse_cluster_v2" "my_cluster" {
  name        = "sharded"
  environment = "PRODUCTION"
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

  shards = {
    "shard1" = {
      weight = 110
    }
    "shard2" = {
      weight = 300
    }
  }

  hosts = {
    "c1a" = {
      type       = "CLICKHOUSE"
      zone       = "ru-central1-a"
      subnet_id  = yandex_vpc_subnet.foo.id
      shard_name = "shard1"
    }
    "c1b" = {
      type       = "CLICKHOUSE"
      zone       = "ru-central1-b"
      subnet_id  = yandex_vpc_subnet.bar.id
      shard_name = "shard1"
    }
    "c2a" = {
      type       = "CLICKHOUSE"
      zone       = "ru-central1-b"
      subnet_id  = yandex_vpc_subnet.bar.id
      shard_name = "shard2"
    }
    "c2d" = {
      type       = "CLICKHOUSE"
      zone       = "ru-central1-d"
      subnet_id  = yandex_vpc_subnet.baz.id
      shard_name = "shard2"
    }
  }

  shard_group {
    name        = "single_shard_group"
    description = "Cluster configuration that contain only shard1"
    shard_names = [
      "shard1",
    ]
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
