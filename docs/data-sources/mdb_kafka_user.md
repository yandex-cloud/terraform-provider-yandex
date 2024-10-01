---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: yandex_mdb_kafka_user"
description: |-
  Get information about a user of the Yandex Managed Kafka cluster.
---


# yandex_mdb_kafka_user




Get information about a user of the Yandex Managed Kafka cluster. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example usage

```terraform
data "yandex_mdb_kafka_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
  password   = "pass123"
}

output "username" {
  value = data.yandex_mdb_kafka_user.foo.name
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kafka cluster.
* `name` - (Required) The name of the Kafka user.
* `password` - (Required) The password of the user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `topic_name` - (Required) The name of the topic that the permission grants access to.
* `role` - (Required) The role type to grant to the topic.
* `allow_hosts` - (Optional) Set of hosts, to which this permission grants access to.
