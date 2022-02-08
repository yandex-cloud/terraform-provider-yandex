---
layout: "yandex"
page_title: "Yandex: yandex_mdb_kafka_cluster"
sidebar_current: "docs-yandex-mdb-kafka-cluster"
description: |-
  Manages a Kafka cluster within Yandex.Cloud.
---

# yandex\_mdb\_kafka\_cluster

Manages a Kafka cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example Usage

Example of creating a Single Node Kafka.

```hcl
resource "yandex_mdb_kafka_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  subnet_ids  = ["${yandex_vpc_subnet.foo.id}"]

  config {
    version          = "2.8"
    brokers_count    = 1
    zones            = ["ru-central1-a"]
    assign_public_ip = false
    unmanaged_topics = false
    schema_registry  = false
    kafka {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 32
      }
      kafka_config {
        compression_type                = "COMPRESSION_TYPE_ZSTD"
        log_flush_interval_messages     = 1024
        log_flush_interval_ms           = 1000
        log_flush_scheduler_interval_ms = 1000
        log_retention_bytes             = 1073741824
        log_retention_hours             = 168
        log_retention_minutes           = 10080
        log_retention_ms                = 86400000
        log_segment_bytes               = 134217728
        log_preallocate                 = true 
        num_partitions                  = 10
        default_replication_factor      = 1 
      }
    }
  }

  user {
    name     = "producer-application"
    password = "password"
    permission {
      topic_name = "input"
      role = "ACCESS_ROLE_PRODUCER"
    }
  }

  user {
    name     = "worker"
    password = "password"
    permission {
      topic_name = "input"
      role = "ACCESS_ROLE_CONSUMER"
    }
    permission {
      topic_name = "output"
      role = "ACCESS_ROLE_PRODUCER"
    }
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a HA Kafka Cluster with two brokers per AZ (6 brokers + 3 zk)

```hcl
resource "yandex_mdb_kafka_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  subnet_ids  = ["${yandex_vpc_subnet.foo.id}", "${yandex_vpc_subnet.bar.id}", "${yandex_vpc_subnet.baz.id}"]

  config {
    version          = "2.8"
    brokers_count    = 2
    zones            = ["ru-central1-a", "ru-central1-b", "ru-central1-c"]
    assign_public_ip = true
    unmanaged_topics = false
    schema_registry  = false
    kafka {
      resources {
        resource_preset_id = "s2.medium"
        disk_type_id       = "network-ssd"
        disk_size          = 128
      }
      kafka_config {
        compression_type                = "COMPRESSION_TYPE_ZSTD"
        log_flush_interval_messages     = 1024
        log_flush_interval_ms           = 1000
        log_flush_scheduler_interval_ms = 1000
        log_retention_bytes             = 1073741824
        log_retention_hours             = 168
        log_retention_minutes           = 10080
        log_retention_ms                = 86400000
        log_segment_bytes               = 134217728
        log_preallocate                 = true
        num_partitions                  = 10
        default_replication_factor      = 6 
      }
    }
    zookeeper {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 20
      }
    }
  }

  user {
    name     = "producer-application"
    password = "password"
    permission {
      topic_name = "input"
      role = "ACCESS_ROLE_PRODUCER"
    }
  }

  user {
    name     = "worker"
    password = "password"
    permission {
      topic_name = "input"
      role = "ACCESS_ROLE_CONSUMER"
    }
    permission {
      topic_name = "output"
      role = "ACCESS_ROLE_PRODUCER"
    }
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Kafka cluster. Provided by the client when the cluster is created.

* `description` - (Optional) Description of the Kafka cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Kafka cluster.

* `network_id` - (Required) ID of the network, to which the Kafka cluster belongs.

* `subnet_ids` - (Optional) IDs of the subnets, to which the Kafka cluster belongs.

* `environment` - (Optional) Deployment environment of the Kafka cluster. Can be either `PRESTABLE` or `PRODUCTION`. 
  The default is `PRODUCTION`.

* `config` - (Required) Configuration of the Kafka cluster. The structure is documented below.

* `user` - (Optional) A user of the Kafka cluster. The structure is documented below.

* `topic` - (Deprecated) To manage topics, please switch to using a separate resource type `yandex_mdb_kafka_topic`.

* `security_group_ids` - (Optional) Security group ids, to which the Kafka cluster belongs.

* `host_group_ids` - (Optional) A list of IDs of the host groups to place VMs of the cluster on.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.

* `maintenance_window` - (Optional) Maintenance policy of the Kafka cluster. The structure is documented below.

~> **Note:** Historically, `topic` blocks of the `yandex_mdb_kafka_cluster` resource were used to manage topics of the Kafka cluster.
However, this approach has a number of disadvantages. In particular, when adding and removing topics from the tf recipe,
terraform generates a diff that misleads the user about the planned changes. Also, this approach turned out to be
inconvenient when managing topics through the Kafka Admin API. Therefore, topic management through a separate resource
type `yandex_mdb_kafka_topic` was implemented and is now recommended.

- - -

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.

* `day` - (Optional) Day of the week (in `DDD` format). Allowed values: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"

* `hour` - (Optional) Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.

- - -

The `config` block supports:

* `version` - (Required) Version of the Kafka server software.

* `brokers_count` - (Optional) Count of brokers per availability zone. The default is `1`.

* `zones` - (Required) List of availability zones.

* `assign_public_ip` - (Optional) Determines whether each broker will be assigned a public IP address. The default is `false`.

* `unmanaged_topics` - (Optional) Allows to use Kafka AdminAPI to manage topics. The default is `false`.

* `schema_registry` - (Optional) Enables managed schema registry on cluster. The default is `false`.

* `kafka` - (Optional) Configuration of the Kafka subcluster. The structure is documented below.

* `zookeeper` - (Optional) Configuration of the ZooKeeper subcluster. The structure is documented below.

- - -

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

The `topic` block is deprecated. To manage topics, please switch to using a separate resource type
`yandex_mdb_kafka_topic`. The `topic` block supports:

* `name` - (Required) The name of the topic.

* `partitions` - (Required) The number of the topic's partitions.

* `replication_factor` - (Required) Amount of data copies (replicas) for the topic in the cluster.

* `topic_config` - (Required) User-defined settings for the topic. The structure is documented below.

The `topic_config` block supports:

* `compression_type`, `delete_retention_ms`, `file_delete_delay_ms`, `flush_messages`, `flush_ms`, `min_compaction_lag_ms`,
`retention_bytes`, `retention_ms`, `max_message_bytes`, `min_insync_replicas`, `segment_bytes`, `preallocate`, - (Optional) Kafka topic settings. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/operations/cluster-topics#update-topic)
and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-kafka/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-kafka/api-ref/Cluster/).

* `host` - A host of the Kafka cluster. The structure is documented below.

The `host` block supports:

* `name` - The fully qualified domain name of the host.
* `zone_id` - The availability zone where the Kafka host was created.
* `role` - Role of the host in the cluster.
* `health` - Health of the host.
* `subnet_id` - The ID of the subnet, to which the host belongs.
* `assign_public_ip` - The flag that defines whether a public IP address is assigned to the node.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_kafka_cluster.foo cluster_id
```
