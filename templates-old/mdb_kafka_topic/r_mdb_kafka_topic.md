---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a topic of a Kafka cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a topic of a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

## Example usage

{{ tffile "examples/mdb_kafka_topic/r_mdb_kafka_topic_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the topic.

* `partitions` - (Required) The number of the topic's partitions.

* `replication_factor` - (Required) Amount of data copies (replicas) for the topic in the cluster.

* `topic_config` - (Optional) User-defined settings for the topic. The structure is documented below.

The `topic_config` block supports:

* `cleanup_policy`, `compression_type`, `delete_retention_ms`, `file_delete_delay_ms`, `flush_messages`, `flush_ms`, `min_compaction_lag_ms`, `retention_bytes`, `retention_ms`, `max_message_bytes`, `min_insync_replicas`, `segment_bytes` - (Optional) Kafka topic settings. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/settings-list#topic-settings) and [the Kafka documentation](https://kafka.apache.org/documentation/#topicconfigs).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_kafka_topic/import.sh" }}
