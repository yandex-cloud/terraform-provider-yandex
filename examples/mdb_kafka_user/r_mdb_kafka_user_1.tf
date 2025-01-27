resource "yandex_mdb_kafka_cluster" "foo" {
  name       = "foo"
  network_id = "c64vs98keiqc7f24pvkd"

  config {
    version = "2.8"
    zones   = ["ru-central1-a"]
    kafka {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-hdd"
        disk_size          = 16
      }
    }
  }
}

resource "yandex_mdb_kafka_topic" "events" {
  cluster_id         = yandex_mdb_kafka_cluster.foo.id
  name               = "events"
  partitions         = 4
  replication_factor = 1
}

resource "yandex_mdb_kafka_user" "user_events" {
  cluster_id = yandex_mdb_kafka_cluster.foo.id
  name       = "user-events"
  password   = "pass1231232332"
  permission {
    topic_name  = "events"
    role        = "ACCESS_ROLE_CONSUMER"
    allow_hosts = ["host1.db.yandex.net", "host2.db.yandex.net"]
  }
  permission {
    topic_name = "events"
    role       = "ACCESS_ROLE_PRODUCER"
  }
}
