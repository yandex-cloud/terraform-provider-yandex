//
// Create a new YDB Dedicated Database with auto-scale policy.
//
resource "yandex_ydb_database_dedicated" "database1" {
  name      = "test-ydb-dedicated"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id

  network_id = yandex_vpc_network.my-inst-group-network.id
  subnet_ids = ["${yandex_vpc_subnet.my-inst-group-subnet.id}"]

  resource_preset_id  = "medium"
  deletion_protection = true

  scale_policy {
    auto_scale {
      min_size = 2
      max_size = 8
      target_tracking {
        cpu_utilization_percent = 70
      }
    }
  }

  labels = {
    # enable preview feature
    enable_autoscaling = "1"
  }

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  location {
    region {
      id = "ru-central1"
    }
  }
}
