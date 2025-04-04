//
// Create a new MDB Sharded Clickhouse Cluster.
//
resource "yandex_mdb_clickhouse_cluster" "my_cluster" {
  name        = "sharded"
  environment = "PRODUCTION"
  network_id  = yandex_vpc_network.foo.id

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "password"
    permission {
      database_name = "db_name"
    }
    settings {
      max_memory_usage_for_user               = 1000000000
      read_overflow_mode                      = "throw"
      output_format_json_quote_64bit_integers = true
    }
    quota {
      interval_duration = 3600000
      queries           = 10000
      errors            = 1000
    }
    quota {
      interval_duration = 79800000
      queries           = 50000
      errors            = 5000
    }
  }

  shard {
    name   = "shard1"
    weight = 110
  }

  shard {
    name   = "shard2"
    weight = 300
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-a"
    subnet_id  = yandex_vpc_subnet.foo.id
    shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = yandex_vpc_subnet.bar.id
    shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = yandex_vpc_subnet.bar.id
    shard_name = "shard2"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-d"
    subnet_id  = yandex_vpc_subnet.baz.id
    shard_name = "shard2"
  }

  shard_group {
    name        = "single_shard_group"
    description = "Cluster configuration that contain only shard1"
    shard_names = [
      "shard1",
    ]
  }

  cloud_storage {
    enabled = false
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
