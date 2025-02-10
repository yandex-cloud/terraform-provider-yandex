---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a user of the Yandex Managed Kafka cluster.
---

# {{.Name}} ({{.Type}})

Get information about a user of the Yandex Managed Kafka cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

## Example usage

{{ tffile "examples/mdb_kafka_user/d_mdb_kafka_user_1.tf" }}

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
