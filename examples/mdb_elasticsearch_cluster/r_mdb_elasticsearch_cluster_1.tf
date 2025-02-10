//
// Create a new MDB ElasticSearch Cluster.
//
resource "yandex_mdb_elasticsearch_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {

    admin_password = "super-password"

    data_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 100
      }
    }

  }

  host {
    name             = "node"
    zone             = "ru-central1-a"
    type             = "DATA_NODE"
    assign_public_ip = true
    subnet_id        = yandex_vpc_subnet.foo.id
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
