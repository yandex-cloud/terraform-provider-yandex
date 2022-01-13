---
layout: "yandex"
page_title: "Yandex: yandex_mdb_kafka_cluster"
sidebar_current: "docs-yandex-datasource-mdb-kafka-cluster"
description: |-
  Get information about a Yandex Managed Kafka cluster.
---

# yandex\_mdb\_kafka\_cluster

Get information about a Yandex Managed Kafka cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example Usage

```hcl
data "yandex_mdb_kafka_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_kafka_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Kafka cluster.

* `name` - (Optional) The name of the Kafka cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the Kafka cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the Kafka cluster.
* `labels` - A set of key/value label pairs to assign to the Kafka cluster.
* `environment` - Deployment environment of the Kafka cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `config` - Configuration of the Kafka cluster. The structure is documented below.
* `user` - A user of the Kafka cluster. The structure is documented below.
* `topic` - A topic of the Kafka cluster. The structure is documented below.
* `host` - A host of the Kafka cluster. The structure is documented below.
* `security_group_ids` - A list of security groups IDs of the Kafka cluster.
* `host_group_ids` - A list of IDs of the host groups hosting VMs of the cluster.
* `maintenance_window` - Maintenance window settings of the Kafka cluster. The structure is documented below.

The `config` block supports:

* `version` - (Required) Version of the Kafka server software.
* `brokers_count` - (Optional) Count of brokers per availability zone.
* `zones` - (Optional) List of availability zones.
* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address on creation. Can be either `true` or `false`.
* `unmanaged_topics` - (Optional) Allows to use Kafka AdminAPI to manage topics. Can be either `true` or `false`.
* `schema_registry` - (Optional) Enables managed schema registry on cluster. Can be either `true` or `false`.
* `kafka` - (Optional) Configuration of the Kafka subcluster. The structure is documented below.
* `zookeeper` - (Optional) Configuration of the ZooKeeper subcluster. The structure is documented below.

The `kafka` block supports:

* `resources` - (Required) Resources allocated to hosts of the Kafka subcluster. The structure is documented below.
* `kafka_config` - (Optional) User-defined settings for the Kafka cluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a Kafka host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).
* `disk_size` - (Required) Volume of the storage available to a Kafka host, in gigabytes.
* `disk_type_id` - (Required) Type of the storage of Kafka hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts/storage).

The `kafka_config` block supports:

* `compression_type`, `log_flush_interval_messages`, `log_flush_interval_ms`, `log_flush_scheduler_interval_ms`, `log_retention_bytes`, `log_retention_hours`,
  `log_retention_minutes`, `log_retention_ms`, `log_segment_bytes`, `log_preallocate`, `socket_send_buffer_bytes`, `socket_receive_buffer_bytes`, `auto_create_topics_enable`,
  `num_partitions`, `default_replication_factor` - (Optional) Kafka server settings. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/operations/cluster-update)
and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration).

The `zookeeper` block supports:

* `resources` - (Optional) Resources allocated to hosts of the ZooKeeper subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Optional) The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).
* `disk_size` - (Optional) Volume of the storage available to a ZooKeeper host, in gigabytes.
* `disk_type_id` - (Optional) Type of the storage of ZooKeeper hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts/storage).

The `user` block supports:

* `name` - (Required) The name of the user.
* `password` - (Required) The password of the user.
* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `topic_name` - (Required) The name of the topic that the permission grants access to.
* `role` - (Required) The role type to grant to the topic.

The `topic` block supports:

* `name` - (Required) The name of the topic.
* `partitions` - (Required) The number of the topic's partitions.
* `replication_factor` - (Required) Amount of data copies (replicas) for the topic in the cluster.
* `topic_config` - (Required) User-defined settings for the topic. The structure is documented below.

The `topic_config` block supports:

* `compression_type`, `delete_retention_ms`, `file_delete_delay_ms`, `flush_messages`, `flush_ms`, `min_compaction_lag_ms`,
`retention_bytes`, `retention_ms`, `max_message_bytes`, `min_insync_replicas`, `segment_bytes`, `preallocate`, - (Optional) Kafka topic settings. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/operations/cluster-topics#update-topic)
and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration).

The `host` block supports:

* `name` - The fully qualified domain name of the host.
* `zone_id` - The availability zone where the Kafka host was created.
* `role` - Role of the host in the cluster.
* `health` - Health of the host.
* `subnet_id` - The ID of the subnet, to which the host belongs.
* `assign_public_ip` - The flag that defines whether a public IP address is assigned to the node.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`.
* `day` - Day of the week (in `DDD` format). Value is one of: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"
* `hour` - Hour of the day in UTC (in `HH` format). Value is between 1 and 24.
