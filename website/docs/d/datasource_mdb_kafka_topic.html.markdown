---
layout: "yandex"
page_title: "Yandex: yandex_mdb_kafka_topic"
sidebar_current: "docs-yandex-datasource-mdb-kafka-topic"
description: |-
  Get information about a topic of the Yandex Managed Kafka cluster.
---

# yandex\_mdb\_kafka\_topic

Get information about a topic of the Yandex Managed Kafka cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example Usage

```hcl
data "yandex_mdb_kafka_topic" "foo" {
  cluster_id = "some_cluster_id"
  name = "test"
}

output "replication_factor" {
  value = "${data.yandex_mdb_kafka_topic.foo.replication_factor}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kafka cluster.
* `name` - (Required) The name of the Kafka topic.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `partitions` - The number of the topic's partitions.
* `replication_factor` - Amount of data copies (replicas) for the topic in the cluster.
* `topic_config` - User-defined settings for the topic. The structure is documented below.

The `topic_config` block supports:

* `cleanup_policy`, `compression_type`, `delete_retention_ms`, `file_delete_delay_ms`, `flush_messages`, `flush_ms`, 
  `min_compaction_lag_ms`, `retention_bytes`, `retention_ms`, `max_message_bytes`, `min_insync_replicas`, 
  `segment_bytes`, `preallocate` - Kafka topic settings. For more information, see
  [the official documentation](https://cloud.yandex.com/en-ru/docs/managed-kafka/concepts/settings-list#topic-settings)
  and [the Kafka documentation](https://kafka.apache.org/documentation/#topicconfigs).
