resource "yandex_kubernetes_cluster" "regional_cluster_resource_name" {
  name        = "name"
  description = "description"

  network_id = yandex_vpc_network.network_resource_name.id

  master {
    regional {
      region = "ru-central1"

      location {
        zone      = yandex_vpc_subnet.subnet_a_resource_name.zone
        subnet_id = yandex_vpc_subnet.subnet_a_resource_name.id
      }

      location {
        zone      = yandex_vpc_subnet.subnet_b_resource_name.zone
        subnet_id = yandex_vpc_subnet.subnet_b_resource_name.id
      }

      location {
        zone      = yandex_vpc_subnet.subnet_d_resource_name.zone
        subnet_id = yandex_vpc_subnet.subnet_d_resource_name.id
      }
    }

    version   = "1.30"
    public_ip = true

    maintenance_policy {
      auto_upgrade = true

      maintenance_window {
        day        = "monday"
        start_time = "15:00"
        duration   = "3h"
      }

      maintenance_window {
        day        = "friday"
        start_time = "10:00"
        duration   = "4h30m"
      }
    }

    master_logging {
      enabled                    = true
      folder_id                  = data.yandex_resourcemanager_folder.folder_resource_name.id
      kube_apiserver_enabled     = true
      cluster_autoscaler_enabled = true
      events_enabled             = true
      audit_enabled              = true
    }
  }

  service_account_id      = yandex_iam_service_account.service_account_resource_name.id
  node_service_account_id = yandex_iam_service_account.node_service_account_resource_name.id

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

  release_channel = "STABLE"
}
