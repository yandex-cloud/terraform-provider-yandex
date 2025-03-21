---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a user of a Kafka cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a user of a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

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
* `allow_hosts` - (Optional) Set of hosts, to which this permission grants access to. Only ip-addresses allowed as value of single host.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_kafka_user/import.sh" }}
