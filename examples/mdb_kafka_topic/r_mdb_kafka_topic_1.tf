//
// Create a new MDB Kafka Topic.
//
resource "yandex_mdb_kafka_topic" "events" {
  cluster_id         = yandex_mdb_kafka_cluster.my_cluster.id
  name               = "events"
  partitions         = 4
  replication_factor = 1
  topic_config {
    cleanup_policy        = "CLEANUP_POLICY_COMPACT"
    compression_type      = "COMPRESSION_TYPE_LZ4"
    delete_retention_ms   = 86400000
    file_delete_delay_ms  = 60000
    flush_messages        = 128
    flush_ms              = 1000
    min_compaction_lag_ms = 0
    retention_bytes       = 10737418240
    retention_ms          = 604800000
    max_message_bytes     = 1048588
    min_insync_replicas   = 1
    segment_bytes         = 268435456
    preallocate           = true
  }
}

resource "yandex_mdb_kafka_cluster" "my_cluster" {
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
