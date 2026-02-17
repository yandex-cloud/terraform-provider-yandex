//
// Create a new Spark Cluster.
//
resource "yandex_spark_cluster" "my_spark_cluster" {

  name               = "spark-cluster-1"
  description        = "created by terraform"
  service_account_id = yandex_iam_service_account.for-spark.id

  labels = {
    my_key = "my_value"
  }

  config = {
    spark_version = "3.5.7"

    resource_pools = {
      driver = {
        resource_preset_id = "c2-m8"
        size               = 1
      }
      executor = {
        resource_preset_id = "c4-m16"
        min_size           = 1
        max_size           = 2
      }
    }
    dependencies = {
      pip_packages = ["numpy==2.2.2"]
    }
  }

  network = {
    subnet_ids         = [yandex_vpc_subnet.a.id]
    security_group_ids = [yandex_vpc_security_group.spark-sg1.id]
  }

  logging = {
    enabled   = true
    folder_id = var.folder_id
  }

  maintenance_window = {
    type = "WEEKLY"
    day  = "TUE"
    hour = 10
  }
}
