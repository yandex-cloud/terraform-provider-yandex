//
// Create a new Data Processing Cluster.
//
resource "yandex_dataproc_cluster" "foo" {
  depends_on = [yandex_resourcemanager_folder_iam_binding.dataproc]

  bucket      = yandex_storage_bucket.foo.bucket
  description = "Dataproc Cluster created by Terraform"
  name        = "dataproc-cluster"
  labels = {
    created_by = "terraform"
  }
  service_account_id = yandex_iam_service_account.dataproc.id
  zone_id            = "ru-central1-b"

  cluster_config {
    # Certain cluster version can be set, but better to use default value (last stable version)
    # version_id = "2.0"

    hadoop {
      services = ["HDFS", "YARN", "SPARK", "TEZ", "MAPREDUCE", "HIVE"]
      properties = {
        "yarn:yarn.resourcemanager.am.max-attempts" = 5
      }
      ssh_public_keys = [
      file("~/.ssh/id_rsa.pub")]
      initialization_action {
        uri  = "s3a://yandex_storage_bucket.foo.bucket/scripts/script.sh"
        args = ["arg1", "arg2"]
      }
    }

    subcluster_spec {
      name = "main"
      role = "MASTERNODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 1
    }

    subcluster_spec {
      name = "data"
      role = "DATANODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
    }

    subcluster_spec {
      name = "compute"
      role = "COMPUTENODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
    }

    subcluster_spec {
      name = "compute_autoscaling"
      role = "COMPUTENODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
      autoscaling_config {
        max_hosts_count        = 10
        measurement_duration   = 60
        warmup_duration        = 60
        stabilization_duration = 120
        preemptible            = false
        decommission_timeout   = 60
      }
    }
  }
}

// Auxiliary resources for Data Processing Cluster
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_iam_service_account" "dataproc" {
  name        = "dataproc"
  description = "service account to manage Dataproc Cluster"
}

data "yandex_resourcemanager_folder" "foo" {
  folder_id = "some_folder_id"
}

resource "yandex_resourcemanager_folder_iam_binding" "dataproc" {
  folder_id = data.yandex_resourcemanager_folder.foo.id
  role      = "mdb.dataproc.agent"
  members = [
    "serviceAccount:${yandex_iam_service_account.dataproc.id}",
  ]
}

// required in order to create bucket
resource "yandex_resourcemanager_folder_iam_binding" "bucket-creator" {
  folder_id = data.yandex_resourcemanager_folder.foo.id
  role      = "editor"
  members = [
    "serviceAccount:${yandex_iam_service_account.dataproc.id}",
  ]
}

resource "yandex_iam_service_account_static_access_key" "foo" {
  service_account_id = yandex_iam_service_account.dataproc.id
}

resource "yandex_storage_bucket" "foo" {
  depends_on = [
    yandex_resourcemanager_folder_iam_binding.bucket-creator
  ]

  bucket     = "foo"
  access_key = yandex_iam_service_account_static_access_key.foo.access_key
  secret_key = yandex_iam_service_account_static_access_key.foo.secret_key
}
