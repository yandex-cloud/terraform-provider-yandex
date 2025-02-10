//
// Create a new Managed Kubernetes Node Group.
//
resource "yandex_kubernetes_node_group" "my_node_group" {
  cluster_id  = yandex_kubernetes_cluster.my_cluster.id
  name        = "name"
  description = "description"
  version     = "1.30"

  labels = {
    "key" = "value"
  }

  instance_template {
    platform_id = "standard-v2"

    network_interface {
      nat        = true
      subnet_ids = ["${yandex_vpc_subnet.my_subnet.id}"]
    }

    resources {
      memory = 2
      cores  = 2
    }

    boot_disk {
      type = "network-hdd"
      size = 64
    }

    scheduling_policy {
      preemptible = false
    }

    container_runtime {
      type = "containerd"
    }
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    location {
      zone = "ru-central1-a"
    }
  }

  maintenance_policy {
    auto_upgrade = true
    auto_repair  = true

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
}
