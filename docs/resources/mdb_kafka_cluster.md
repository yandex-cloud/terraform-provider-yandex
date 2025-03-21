---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: yandex_mdb_kafka_cluster"
description: |-
  Manages a Kafka cluster within Yandex Cloud.
---

# yandex_mdb_kafka_cluster (Resource)

Manages a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

~> Historically, `topic` blocks of the `yandex_mdb_kafka_cluster` resource were used to manage topics of the Kafka cluster. However, this approach has a number of disadvantages. In particular, when adding and removing topics from the tf recipe, terraform generates a diff that misleads the user about the planned changes. Also, this approach turned out to be inconvenient when managing topics through the Kafka Admin API. Therefore, topic management through a separate resource type `yandex_mdb_kafka_topic` was implemented and is now recommended.

## Example usage

```terraform
//
// Create a new MDB Kafka Cluster.
//
resource "yandex_mdb_kafka_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  subnet_ids  = ["${yandex_vpc_subnet.foo.id}"]

  config {
    version          = "2.8"
    brokers_count    = 1
    zones            = ["ru-central1-a"]
    assign_public_ip = false
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
        message_max_bytes               = 1048588
        replica_fetch_max_bytes         = 1048576
        ssl_cipher_suites               = ["TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"]
        offsets_retention_minutes       = 10080
        sasl_enabled_mechanisms         = ["SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"]
      }
    }
  }

  user {
    name     = "producer-application"
    password = "password"
    permission {
      topic_name  = "input"
      role        = "ACCESS_ROLE_PRODUCER"
      allow_hosts = ["host1.db.yandex.net", "host2.db.yandex.net"]
    }
  }

  user {
    name     = "worker"
    password = "password"
    permission {
      topic_name = "input"
      role       = "ACCESS_ROLE_CONSUMER"
    }
    permission {
      topic_name = "output"
      role       = "ACCESS_ROLE_PRODUCER"
    }
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

```terraform
//
// Create a new MDB HA Kafka Cluster with two brokers per AZ.
// (6 brokers & 3 Zookeepers)
//
resource "yandex_mdb_kafka_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  subnet_ids  = ["${yandex_vpc_subnet.foo.id}", "${yandex_vpc_subnet.bar.id}", "${yandex_vpc_subnet.baz.id}"]

  config {
    version          = "2.8"
    brokers_count    = 2
    zones            = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
    assign_public_ip = true
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
        message_max_bytes               = 1048588
        replica_fetch_max_bytes         = 1048576
        ssl_cipher_suites               = ["TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"]
        offsets_retention_minutes       = 10080
        sasl_enabled_mechanisms         = ["SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"]
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
      topic_name  = "input"
      role        = "ACCESS_ROLE_PRODUCER"
      allow_hosts = ["host1.db.yandex.net", "host2.db.yandex.net"]
    }
  }

  user {
    name     = "worker"
    password = "password"
    permission {
      topic_name = "input"
      role       = "ACCESS_ROLE_CONSUMER"
    }
    permission {
      topic_name = "output"
      role       = "ACCESS_ROLE_PRODUCER"
    }
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

```terraform
//
// Create a new MDB Kafka Cluster with
// KRaft-controller sub-cluster instead of Zookeeper sub-cluster.
//
resource "yandex_mdb_kafka_cluster" "kraft-split" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  subnet_ids  = ["${yandex_vpc_subnet.foo.id}", "${yandex_vpc_subnet.bar.id}", "${yandex_vpc_subnet.baz.id}"]

  config {
    version          = "3.6"
    brokers_count    = 2
    zones            = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
    assign_public_ip = true
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
        message_max_bytes               = 1048588
        replica_fetch_max_bytes         = 1048576
        ssl_cipher_suites               = ["TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"]
        offsets_retention_minutes       = 10080
        sasl_enabled_mechanisms         = ["SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"]
      }
    }
    kraft {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 20
      }
    }
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

```terraform
//
// Creating multi-host Kafka Cluster without sub-cluster of controllers, 
// using KRaft-combine quorum.
//
resource "yandex_mdb_kafka_cluster" "kraft-combine" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  subnet_ids  = ["${yandex_vpc_subnet.foo.id}", "${yandex_vpc_subnet.bar.id}", "${yandex_vpc_subnet.baz.id}"]

  config {
    version          = "3.6"
    brokers_count    = 1
    zones            = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
    assign_public_ip = true
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
        message_max_bytes               = 1048588
        replica_fetch_max_bytes         = 1048576
        ssl_cipher_suites               = ["TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"]
        offsets_retention_minutes       = 10080
        sasl_enabled_mechanisms         = ["SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"]
      }
    }
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `config` (Block List, Min: 1, Max: 1) Configuration of the Kafka cluster. (see [below for nested schema](#nestedblock--config))
- `name` (String) The resource name.
- `network_id` (String) The `VPC Network ID` of subnets which resource attached to.

### Optional

- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `description` (String) The resource description.
- `environment` (String) Deployment environment of the Kafka cluster. Can be either `PRESTABLE` or `PRODUCTION`. The default is `PRODUCTION`.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `host_group_ids` (Set of String) A list of IDs of the host groups to place VMs of the cluster on.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `maintenance_window` (Block List, Max: 1) Maintenance policy of the Kafka cluster. (see [below for nested schema](#nestedblock--maintenance_window))
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `subnet_ids` (List of String) The list of VPC subnets identifiers which resource is attached.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `topic` (Block List, Deprecated) To manage topics, please switch to using a separate resource type `yandex_mdb_kafka_topic`. (see [below for nested schema](#nestedblock--topic))
- `user` (Block Set, Deprecated) To manage users, please switch to using a separate resource type `yandex_mdb_kafka_user`. (see [below for nested schema](#nestedblock--user))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `health` (String) Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-kafka/api-ref/Cluster/).
- `host` (Set of Object) A host of the Kafka cluster. (see [below for nested schema](#nestedatt--host))
- `id` (String) The ID of this resource.
- `status` (String) Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-kafka/api-ref/Cluster/).

<a id="nestedblock--config"></a>
### Nested Schema for `config`

Required:

- `kafka` (Block List, Min: 1, Max: 1) Configuration of the Kafka subcluster. (see [below for nested schema](#nestedblock--config--kafka))
- `version` (String) Version of the Kafka server software.
- `zones` (List of String) List of availability zones.

Optional:

- `access` (Block List, Max: 1) Access policy to the Kafka cluster. (see [below for nested schema](#nestedblock--config--access))
- `assign_public_ip` (Boolean) Determines whether each broker will be assigned a public IP address. The default is `false`.
- `brokers_count` (Number) Count of brokers per availability zone. The default is `1`.
- `disk_size_autoscaling` (Block List, Max: 1) Disk autoscaling settings of the Kafka cluster. (see [below for nested schema](#nestedblock--config--disk_size_autoscaling))
- `kraft` (Block List, Max: 1) Configuration of the KRaft-controller subcluster. (see [below for nested schema](#nestedblock--config--kraft))
- `schema_registry` (Boolean) Enables managed schema registry on cluster. The default is `false`.
- `unmanaged_topics` (Boolean, Deprecated)
- `zookeeper` (Block List, Max: 1) Configuration of the ZooKeeper subcluster. (see [below for nested schema](#nestedblock--config--zookeeper))

<a id="nestedblock--config--kafka"></a>
### Nested Schema for `config.kafka`

Required:

- `resources` (Block List, Min: 1, Max: 1) Resources allocated to hosts of the Kafka subcluster. (see [below for nested schema](#nestedblock--config--kafka--resources))

Optional:

- `kafka_config` (Block List, Max: 1) User-defined settings for the Kafka cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/operations/cluster-update) and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration). (see [below for nested schema](#nestedblock--config--kafka--kafka_config))

<a id="nestedblock--config--kafka--resources"></a>
### Nested Schema for `config.kafka.resources`

Required:

- `disk_size` (Number) Volume of the storage available to a Kafka host, in gigabytes.
- `disk_type_id` (String) Type of the storage of Kafka hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/storage).
- `resource_preset_id` (String) The ID of the preset for computational resources available to a Kafka host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).


<a id="nestedblock--config--kafka--kafka_config"></a>
### Nested Schema for `config.kafka.kafka_config`

Optional:

- `auto_create_topics_enable` (Boolean)
- `compression_type` (String)
- `default_replication_factor` (String)
- `log_flush_interval_messages` (String)
- `log_flush_interval_ms` (String)
- `log_flush_scheduler_interval_ms` (String)
- `log_preallocate` (Boolean)
- `log_retention_bytes` (String)
- `log_retention_hours` (String)
- `log_retention_minutes` (String)
- `log_retention_ms` (String)
- `log_segment_bytes` (String)
- `message_max_bytes` (String)
- `num_partitions` (String)
- `offsets_retention_minutes` (String)
- `replica_fetch_max_bytes` (String)
- `sasl_enabled_mechanisms` (Set of String)
- `socket_receive_buffer_bytes` (String)
- `socket_send_buffer_bytes` (String)
- `ssl_cipher_suites` (Set of String)



<a id="nestedblock--config--access"></a>
### Nested Schema for `config.access`

Optional:

- `data_transfer` (Boolean) Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer).


<a id="nestedblock--config--disk_size_autoscaling"></a>
### Nested Schema for `config.disk_size_autoscaling`

Required:

- `disk_size_limit` (Number) Maximum possible size of disk in bytes.

Optional:

- `emergency_usage_threshold` (Number) Percent of disk utilization. Disk will autoscale immediately, if this threshold reached. Value is between 0 and 100. Default value is 0 (autoscaling disabled). Must be not less then 'planned_usage_threshold' value.
- `planned_usage_threshold` (Number) Percent of disk utilization. During maintenance disk will autoscale, if this threshold reached. Value is between 0 and 100. Default value is 0 (autoscaling disabled).


<a id="nestedblock--config--kraft"></a>
### Nested Schema for `config.kraft`

Optional:

- `resources` (Block List, Max: 1) Resources allocated to hosts of the KRaft-controller subcluster. (see [below for nested schema](#nestedblock--config--kraft--resources))

<a id="nestedblock--config--kraft--resources"></a>
### Nested Schema for `config.kraft.resources`

Optional:

- `disk_size` (Number) Volume of the storage available to a KRaft-controller host, in gigabytes.
- `disk_type_id` (String) Type of the storage of KRaft-controller hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/storage).
- `resource_preset_id` (String) The ID of the preset for computational resources available to a KRaft-controller host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).



<a id="nestedblock--config--zookeeper"></a>
### Nested Schema for `config.zookeeper`

Optional:

- `resources` (Block List, Max: 1) Resources allocated to hosts of the ZooKeeper subcluster. (see [below for nested schema](#nestedblock--config--zookeeper--resources))

<a id="nestedblock--config--zookeeper--resources"></a>
### Nested Schema for `config.zookeeper.resources`

Optional:

- `disk_size` (Number) Volume of the storage available to a ZooKeeper host, in gigabytes.
- `disk_type_id` (String) Type of the storage of ZooKeeper hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/storage).
- `resource_preset_id` (String) The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).




<a id="nestedblock--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Required:

- `type` (String) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.

Optional:

- `day` (String) Day of the week (in `DDD` format). Allowed values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.
- `hour` (Number) Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedblock--topic"></a>
### Nested Schema for `topic`

Required:

- `name` (String) The name of the topic.
- `partitions` (Number) The number of the topic's partitions.
- `replication_factor` (Number) Amount of data copies (replicas) for the topic in the cluster.

Optional:

- `topic_config` (Block List, Max: 1) User-defined settings for the topic. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/operations/cluster-topics#update-topic) and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration). (see [below for nested schema](#nestedblock--topic--topic_config))

<a id="nestedblock--topic--topic_config"></a>
### Nested Schema for `topic.topic_config`

Optional:

- `cleanup_policy` (String)
- `compression_type` (String)
- `delete_retention_ms` (String)
- `file_delete_delay_ms` (String)
- `flush_messages` (String)
- `flush_ms` (String)
- `max_message_bytes` (String)
- `min_compaction_lag_ms` (String)
- `min_insync_replicas` (String)
- `preallocate` (Boolean)
- `retention_bytes` (String)
- `retention_ms` (String)
- `segment_bytes` (String)



<a id="nestedblock--user"></a>
### Nested Schema for `user`

Required:

- `name` (String) The name of the user.
- `password` (String, Sensitive) The password of the user.

Optional:

- `permission` (Block Set) Set of permissions granted to the user. (see [below for nested schema](#nestedblock--user--permission))

<a id="nestedblock--user--permission"></a>
### Nested Schema for `user.permission`

Required:

- `role` (String) The role type to grant to the topic.
- `topic_name` (String) The name of the topic that the permission grants access to.

Optional:

- `allow_hosts` (Set of String) Set of hosts, to which this permission grants access to. Only ip-addresses allowed as value of single host.



<a id="nestedatt--host"></a>
### Nested Schema for `host`

Read-Only:

- `assign_public_ip` (Boolean)
- `health` (String)
- `name` (String)
- `role` (String)
- `subnet_id` (String)
- `zone_id` (String)

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_mdb_kafka_cluster.<resource Name> <resource Id>
terraform import yandex_mdb_kafka_cluster.my_cluster ...
```
