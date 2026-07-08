//
// Create a new MDB MySQL Backup Retention Policy.
//
// The API does not support updates, so changing any argument
// forces the policy to be replaced (destroyed and recreated).
//
resource "yandex_mdb_mysql_backup_retention_policy" "my_policy" {
  cluster_id      = yandex_mdb_mysql_cluster.my_cluster.id
  policy_name     = "keep-weekly-backups"
  description     = "Keep weekly backups for 30 days"
  retain_for_days = 30

  cron = {
    day_of_month = "*"
    day_of_week  = "1"
    month        = "*"
  }
}

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

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
