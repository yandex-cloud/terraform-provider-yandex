//
// Create a new MDB OpenSearch Cluster.
//
resource "yandex_mdb_opensearch_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {

    admin_password = "super-password"

    opensearch {
      node_groups {
        name             = "group0"
        assign_public_ip = true
        hosts_count      = 1
        subnet_ids       = ["${yandex_vpc_subnet.foo.id}"]
        zone_ids         = ["ru-central1-d"]
        roles            = ["data", "manager"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
