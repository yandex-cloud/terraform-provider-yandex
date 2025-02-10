//
// Create a new Compute Instance Group (IG)
//
resource "yandex_compute_instance_group" "group1" {
  name                = "test-ig"
  folder_id           = data.yandex_resourcemanager_folder.test_folder.id
  service_account_id  = yandex_iam_service_account.test_account.id
  deletion_protection = true
  instance_template {
    platform_id = "standard-v1"
    resources {
      memory = 2
      cores  = 2
    }
    boot_disk {
      mode = "READ_WRITE"
      initialize_params {
        image_id = data.yandex_compute_image.ubuntu.id
        size     = 4
      }
    }
    network_interface {
      network_id = yandex_vpc_network.my-inst-group-network.id
      subnet_ids = ["${yandex_vpc_subnet.my-inst-group-subnet.id}"]
    }
    labels = {
      label1 = "label1-value"
      label2 = "label2-value"
    }
    metadata = {
      foo      = "bar"
      ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
    }
    network_settings {
      type = "STANDARD"
    }
  }

  variables = {
    test_key1 = "test_value1"
    test_key2 = "test_value2"
  }

  scale_policy {
    fixed_scale {
      size = 3
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 2
    max_creating    = 2
    max_expansion   = 2
    max_deleting    = 2
  }
}
