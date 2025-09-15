locals {
  shards = 4
  users = {
    alice: {
      "password": "password",
      "conn_limit": 30
    },
    bob: {
      "password": "mysupercoolpassword",
      "conn_limit": 15

    }
  }
  dbs = {"testdb": "alice", "anotherdb": "bob"}

  user_shard_combinations = flatten([
    for shard in range(local.shards) : [
      for user, cfg in local.users : {
        shard = shard
        name  = user
        settings = cfg
        key   = "${shard}-${user}"
      }
    ]
  ])
  db_shard_combinations = flatten([
    for shard in range(local.shards) : [
      for db, owner in local.dbs : {
        shard = shard
        db  = db
        owner = owner
        key   = "${shard}-${db}"
        owner_key = "${shard}-${owner}"
      }
    ]
  ])
}

/****************** PostgreSQL Shards Management ******************/

resource "yandex_mdb_postgresql_cluster_v2" "shard" {
  count = local.shards

  name        = "shard${count.index}"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 17
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }

    postgresql_config = {
      max_connections                = 395
      enable_parallel_hash           = true
      autovacuum_vacuum_scale_factor = 0.34
      default_transaction_isolation  = "TRANSACTION_ISOLATION_READ_COMMITTED"
      shared_preload_libraries       = "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
      max_wal_senders = 20
      shared_buffers = 2147483648
    }
  }

  hosts = {
    "first" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.test-subnet-a.id
    }
    "second" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.test-subnet-b.id
    }
    "third" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.test-subnet-d.id
    }
  }
}

resource "yandex_mdb_postgresql_user" "shard_user" {
  for_each = { for combo in local.user_shard_combinations : combo.key => combo }

  cluster_id = yandex_mdb_postgresql_cluster_v2.shard[each.value.shard].id
  name       = "${each.value.name}"
  password   = "${each.value.settings.password}"
  conn_limit = "${each.value.settings.conn_limit}"
  settings = {
    default_transaction_isolation = "read committed"
    log_min_duration_statement    = 5000
  }
}

resource "yandex_mdb_postgresql_database" "shard_db" {
  for_each = { for combo in local.db_shard_combinations : combo.key => combo }

  cluster_id = yandex_mdb_postgresql_cluster_v2.shard[each.value.shard].id
  name       = "${each.value.db}"
  owner      = yandex_mdb_postgresql_user.shard_user[each.value.owner_key].name
  lc_collate = "en_US.UTF-8"
  lc_type    = "en_US.UTF-8"
  extension {
    name = "uuid-ossp"
  }
  extension {
    name = "xml2"
  }
}

/****************** Sharded PostgreSQL Cluster Management ******************/

resource "yandex_mdb_sharded_postgresql_cluster" "default" {
  name        = "diphantxm-full"
  environment = "PRODUCTION"
  network_id  = yandex_vpc_network.foo.id

  hosts = {
    "router2" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.test-subnet-a.id
      assign_public_ip = false
      type = "ROUTER"
    }
    "router1" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.test-subnet-b.id
      assign_public_ip = false
      type = "ROUTER"
    }
    "router3" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.test-subnet-d.id
      assign_public_ip = false
      type = "ROUTER"
    }
  }

  config = {
    backup_retain_period_days = 10
    sharded_postgresql_config = {
        common = {
            console_password = "password"
            log_level = "INFO"
        }
        router = {
            resources = {
                resource_preset_id = "s2.micro"
                disk_type_id       = "network-ssd"
                disk_size          = 32
            }
            config = {
              show_notice_messages = false
              prefer_same_availability_zone = true
            }
        }
        balancer = {}
    }
  }
}

resource "yandex_mdb_sharded_postgresql_database" "spqr_db" {
  for_each = local.dbs

  cluster_id = yandex_mdb_sharded_postgresql_cluster.default.id
  name       = "${each.key}"
}

resource "yandex_mdb_sharded_postgresql_user" "spqr_user" {
  for_each = local.users

  cluster_id = yandex_mdb_sharded_postgresql_cluster.default.id
  name       = "${each.key}"
  password   = "${each.value.password}"
  settings = {
    connection_limit = "${each.value.conn_limit}"
    connection_retries = 5
  }
}

resource "yandex_mdb_sharded_postgresql_shard" "shard" {
  count = local.shards

	cluster_id = yandex_mdb_sharded_postgresql_cluster.default.id
	name       = "shard${count.index}"
	shard_spec = {
		mdb_postgresql = yandex_mdb_postgresql_cluster_v2.shard[count.index].id
	}
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
resource "yandex_vpc_subnet" "test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}
resource "yandex_vpc_subnet" "test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
