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

  # resource_groups = file("resource-groups.json")
  resource_groups = jsonencode(
    {
      "rootGroups" : [
        {
          "name" : "global",
          "softMemoryLimit" : "80%",
          "hardConcurrencyLimit" : 100,
          "maxQueued" : 1000,
          "schedulingPolicy" : "weighted",
          "subGroups" : [
            {
              "name" : "adhoc",
              "softMemoryLimit" : "10%",
              "hardConcurrencyLimit" : 50,
              "maxQueued" : 1,
              "schedulingWeight" : 10,
              "subGroups" : [
                {
                  "name" : "other",
                  "softMemoryLimit" : "10%",
                  "hardConcurrencyLimit" : 2,
                  "maxQueued" : 1,
                  "schedulingWeight" : 10,
                  "schedulingPolicy" : "weighted_fair",
                  "subGroups" : [
                    {
                      "name" : "$${USER}",
                      "softMemoryLimit" : "10%",
                      "hardConcurrencyLimit" : 1,
                      "maxQueued" : 100
                    }
                  ]
                }
              ]
            }
          ]
        },
        {
          "name" : "admin",
          "softMemoryLimit" : "100%",
          "hardConcurrencyLimit" : 50,
          "maxQueued" : 100,
          "schedulingPolicy" : "query_priority"
        }
      ],
      "selectors" : [
        {
          "user" : "bob",
          "userGroup" : "admin",
          "queryType" : "DATA_DEFINITION",
          "source" : "jdbc#(?<toolname>.*)",
          "clientTags" : [
            "hipri"
          ],
          "group" : "admin"
        },
        {
          "group" : "global.adhoc.other.$${USER}"
        }
      ],
      "cpuQuotaPeriod" : "1h"
    }
  )


  query_properties = {
    "query.max-memory-per-node"     = "7GB"
    "memory.heap-headroom-per-node" = "31%"
    "query.max-memory"              = "13GB"
    "query.max-total-memory"        = "21GB"
  }

  maintenance_window = {
    day  = "MON"
    hour = 15
    type = "ANYTIME"
  }
}
