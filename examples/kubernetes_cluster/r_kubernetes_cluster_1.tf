//
// Create a new Managed Kubernetes zonal Cluster.
//
resource "yandex_kubernetes_cluster" "zonal_cluster" {
  name        = "name"
  description = "description"

  network_id = yandex_vpc_network.network_resource_name.id

  master {
    version = "1.30"
    zonal {
      zone      = yandex_vpc_subnet.subnet_resource_name.zone
      subnet_id = yandex_vpc_subnet.subnet_resource_name.id
    }

    public_ip = true

    security_group_ids = ["${yandex_vpc_security_group.security_group_name.id}"]

    maintenance_policy {
      auto_upgrade = true

      maintenance_window {
        start_time = "15:00"
        duration   = "3h"
      }
    }

    master_logging {
      enabled                    = true
      log_group_id               = yandex_logging_group.log_group_resource_name.id
      kube_apiserver_enabled     = true
      cluster_autoscaler_enabled = true
      events_enabled             = true
      audit_enabled              = true
    }

    scale_policy {
      auto_scale {
        min_resource_preset_id = "s-c4-m16"
      }
    }
  }

  service_account_id      = yandex_iam_service_account.service_account_resource_name.id
  node_service_account_id = yandex_iam_service_account.node_service_account_resource_name.id

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

  release_channel         = "RAPID"
  network_policy_provider = "CALICO"

  kms_provider {
    key_id = yandex_kms_symmetric_key.kms_key_resource_name.id
  }

  workload_identity_federation {
    enabled = true
  }
}
