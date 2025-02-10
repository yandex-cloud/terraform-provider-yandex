//
// Create a new MDB Sharded Redis Cluster.
//
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  sharded     = true

  config {
    version  = "6.2"
    password = "your_password"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone       = "ru-central1-a"
    subnet_id  = yandex_vpc_subnet.foo.id
    shard_name = "first"
  }

  host {
    zone       = "ru-central1-b"
    subnet_id  = yandex_vpc_subnet.bar.id
    shard_name = "second"
  }

  host {
    zone       = "ru-central1-d"
    subnet_id  = yandex_vpc_subnet.baz.id
    shard_name = "third"
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
