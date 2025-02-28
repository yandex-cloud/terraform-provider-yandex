resource "yandex_mdb_redis_cluster_v2" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config = {
    password = "your_password"
    version  = "7.2-valkey"
  }

  resources = {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  hosts = {
      "aaa" = {
        zone      = "ru-central1-a"
        subnet_id = yandex_vpc_subnet.foo.id
      }
  }

  maintenance_window = {
    type = "ANYTIME"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
