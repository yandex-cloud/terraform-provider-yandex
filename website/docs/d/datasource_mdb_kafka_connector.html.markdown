---
layout: "yandex"
page_title: "Yandex: yandex_mdb_kafka_connector"
sidebar_current: "docs-yandex-datasource-mdb-kafka-connector"
description: |-
  Get information about a connector of the Yandex Managed Kafka cluster.
---

# yandex\_mdb\_kafka\_connector

Get information about a connector of the Yandex Managed Kafka cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).

## Example Usage

```hcl
data "yandex_mdb_kafka_connector" "foo" {
  cluster_id = "some_cluster_id"
  name = "test"
}

output "tasks_max" {
  value = "${data.yandex_mdb_kafka_connector.foo.tasks_max}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kafka cluster.
* `name` - (Required) The name of the Kafka connector.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `partitions` - The number of the topic's partitions.
* `tasks_max` - The number of the connector's parallel working tasks. Default is the number of brokers
* `properties` - Additional properties for connector.
* `connector_config_mirrormaker` - Params for MirrorMaker2 connector. The structure is documented below.

The `connector_config_mirrormaker` block supports:
* `topics` - The pattern for topic names to be replicated.
* `replication_factor` - Replication factor for topics created in target cluster
* `source_cluster` - Settings for source cluster. The structure is documented below.
* `target_cluster` - Settings for target cluster. The structure is documented below.

The `source_cluster` and `target_cluster` block supports:
* `alias` - Name of the cluster. Used also as a topic prefix
* `external_cluster` - Connection params for external cluster
* `this_cluster` - Using this section in the cluster definition (source or target) means it's this cluster

The `external_cluster` blocks support:
* `bootstrap_servers` - List of bootstrap servers to connect to cluster
* `sasl_username` - Username to use in SASL authentification mechanism
* `sasl_password` - Password to use in SASL authentification mechanism
* `sasl_mechanism` - Type of SASL authentification mechanism to use
* `security_protocol` - Security protocol to use
