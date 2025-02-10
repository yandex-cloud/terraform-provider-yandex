---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: yandex_mdb_kafka_topic"
description: |-
  Manages a topic of a Kafka cluster within Yandex Cloud.
---

# yandex_mdb_kafka_topic (Resource)

Manages a topic of a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

## Example usage

```terraform
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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the topic.

* `partitions` - (Required) The number of the topic's partitions.

* `replication_factor` - (Required) Amount of data copies (replicas) for the topic in the cluster.

* `topic_config` - (Optional) User-defined settings for the topic. The structure is documented below.

The `topic_config` block supports:

* `cleanup_policy`, `compression_type`, `delete_retention_ms`, `file_delete_delay_ms`, `flush_messages`, `flush_ms`, `min_compaction_lag_ms`, `retention_bytes`, `retention_ms`, `max_message_bytes`, `min_insync_replicas`, `segment_bytes`, `preallocate` - (Optional) Kafka topic settings. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/settings-list#topic-settings) and [the Kafka documentation](https://kafka.apache.org/documentation/#topicconfigs).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_mdb_kafka_topic.<resource Name> <resource Id>
terraform import yandex_mdb_kafka_topic.events ...
```
