
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "test_ha"
  description = "test High-Availability (HA) PostgreSQL Cluster"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 15
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }

  }

  host {
    zone      = "ru-central1-a"
    name      = "host_name_a"
    subnet_id = yandex_vpc_subnet.a.id
  }
  host {
    zone                    = "ru-central1-b"
    name                    = "host_name_b"
    replication_source_name = "host_name_c"
    subnet_id               = yandex_vpc_subnet.b.id
  }
  host {
    zone      = "ru-central1-c"
    name      = "host_name_c"
    subnet_id = yandex_vpc_subnet.c.id
  }
  host {
    zone      = "ru-central1-c"
    name      = "host_name_c_2"
    subnet_id = yandex_vpc_subnet.c.id
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "c" {
  zone           = "ru-central1-c"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
