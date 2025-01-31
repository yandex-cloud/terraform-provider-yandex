---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a user of a Kafka cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a user of a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example usage

{{ tffile "examples/mdb_kafka_user/r_mdb_kafka_user_1.tf" }}

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
