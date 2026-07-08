//
// Create a new MDB Redis Backup Retention Policy.
//
// The API does not support updates, so changing any argument
// forces the policy to be replaced (destroyed and recreated).
//
resource "yandex_mdb_redis_backup_retention_policy" "my_policy" {
  cluster_id      = yandex_mdb_redis_cluster.my_cluster.id
  policy_name     = "keep-weekly-backups"
  description     = "Keep weekly backups for 30 days"
  retain_for_days = 30

  cron = {
    day_of_month = "*"
    day_of_week  = "1"
    month        = "*"
  }
}

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
    zone       = "ru-central1-d"
    subnet_id  = yandex_vpc_subnet.baz.id
    shard_name = "third"
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
