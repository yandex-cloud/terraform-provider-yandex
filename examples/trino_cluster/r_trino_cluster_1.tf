resource "yandex_trino_cluster" "trino" {
  name               = "trino-created-with-terraform"
  service_account_id = yandex_iam_service_account.trino.id
  subnet_ids         = [yandex_vpc_subnet.a.id, yandex_vpc_subnet.b.id, yandex_vpc_subnet.d.id]

  coordinator = {
    resource_preset_id = "c8-m32"
  }

  worker = {
    fixed_scale = {
      count = 4
    }
    resource_preset_id = "c4-m16"
  }


  retry_policy = {
    additional_properties = {
      fault-tolerant-execution-max-task-split-count = 1024
    }
    policy = "TASK"
    exchange_manager = {
      additional_properties = {
        exchange.sink-buffer-pool-min-size = 16
      }
      service_s3 = {}
    }
  }

  maintenance_window = {
    day  = "MON"
    hour = 15
    type = "ANYTIME"
  }
}
