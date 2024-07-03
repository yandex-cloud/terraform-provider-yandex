---
layout: "yandex"
page_title: "Yandex: yandex_mdb_kafka_user"
sidebar_current: "docs-yandex-datasource-mdb-kafka-user"
description: |-
  Get information about a user of the Yandex Managed Kafka cluster.
---

# yandex\_mdb\_kafka\_user

Get information about a user of the Yandex Managed Kafka cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example Usage

```hcl
data "yandex_mdb_kafka_user" "foo" {
  cluster_id = "some_cluster_id"
  name = "test"
}

output "username" {
  value = "${data.yandex_mdb_kafka_user.foo.name}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kafka cluster.
* `name` - (Required) The name of the Kafka user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `permission` - Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `topic_name` - The name of the topic that the permission grants access to.
* `role` - The role type to grant to the topic.
* `allow_hosts` - Set of hosts, to which this permission grants access to.