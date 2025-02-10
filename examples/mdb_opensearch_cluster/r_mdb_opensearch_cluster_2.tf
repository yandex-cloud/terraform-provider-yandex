//
// Create a new MDB OpenSearch Cluster.
//
locals {
  zones = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
}

resource "yandex_mdb_opensearch_cluster" "my_cluster" {
  name        = "my-cluster"
  environment = "PRODUCTION"
  network_id  = yandex_vpc_network.es-net.id

  config {

    admin_password = "super-password"

    opensearch {
      node_groups {
        name             = "hot_group0"
        assign_public_ip = true
        hosts_count      = 2
        zone_ids         = local.zones
        roles            = ["data"]
        resources {
          resource_preset_id = "s2.small"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }

      node_groups {
        name             = "cold_group0"
        assign_public_ip = true
        hosts_count      = 2
        zone_ids         = local.zones
        roles            = ["data"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-hdd"
        }
      }

      node_groups {
        name             = "managers_group"
        assign_public_ip = true
        hosts_count      = 3
        zone_ids         = local.zones
        roles            = ["manager"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }

      plugins = ["analysis-icu"]
    }

    dashboards {
      node_groups {
        name             = "dashboards"
        assign_public_ip = true
        hosts_count      = 1
        zone_ids         = local.zones
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }
  }

  auth_settings = {
    saml = {
      idp_entity_id             = "urn:dev.auth0.example.com"
      idp_metadata_file_content = "<EntityDescriptor entityID=\"https://test_identity_provider.example.com\"></EntityDescriptor>"
      sp_entity_id              = "https://test.example.com",
      dashboards_url            = "https://dashboards.example.com"
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
