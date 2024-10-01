resource "yandex_loadtesting_agent" "my-agent" {
  name        = "my-agent"
  description = "2 core 4 GB RAM agent"
  folder_id   = data.yandex_resourcemanager_folder.test_folder.id
  labels = {
    jmeter = "5"
  }

  compute_instance {
    zone_id            = "ru-central1-b"
    service_account_id = yandex_iam_service_account.test_account.id
    resources {
      memory = 4
      cores  = 2
    }
    boot_disk {
      initialize_params {
        size = 15
      }
      auto_delete = true
    }
    network_interface {
      subnet_id = yandex_vpc_subnet.my-subnet-a.id
    }
  }
}
