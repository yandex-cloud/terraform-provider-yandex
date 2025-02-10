//
// Create a new MDB ElasticSearch Cluster.
//

locals {
  zones = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
}

resource "yandex_mdb_elasticsearch_cluster" "my_cluster" {
  name        = "my-cluster"
  environment = "PRODUCTION"
  network_id  = yandex_vpc_network.es-net.id

  config {
    edition        = "platinum"
    admin_password = "super-password"

    data_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 100
      }
    }

    master_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 10
      }
    }

    plugins = ["analysis-icu"]

  }

  dynamic "host" {
    for_each = toset(range(0, 6))
    content {
      name             = "datanode${host.value}"
      zone             = local.zones[(host.value) % 3]
      type             = "DATA_NODE"
      assign_public_ip = true
    }
  }

  dynamic "host" {
    for_each = toset(range(0, 3))
    content {
      name = "masternode${host.value}"
      zone = local.zones[host.value % 3]
      type = "MASTER_NODE"
    }
  }

  depends_on = [
    yandex_vpc_subnet.es-subnet-a,
    yandex_vpc_subnet.es-subnet-b,
    yandex_vpc_subnet.es-subnet-d,
  ]

}

// Auxiliary resources
resource "yandex_vpc_network" "es-net" {}

resource "yandex_vpc_subnet" "es-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.es-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "es-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.es-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "es-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.es-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
