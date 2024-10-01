---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: yandex_mdb_kafka_user"
description: |-
  Manages a user of a Kafka cluster within Yandex.Cloud.
---


# yandex_mdb_kafka_user




Manages a user of a Kafka cluster within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example usage

```terraform
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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `topic_name` - (Required) The name of the topic that the permission grants access to.

* `role` - (Required) The role type to grant to the topic.

* `allow_hosts` - (Optional) Set of hosts, to which this permission grants access to.

## Import

Kafka user can be imported using following format:

```
$ terraform import yandex_mdb_kafka_user.foo {cluster_id}:{user_name}
```
