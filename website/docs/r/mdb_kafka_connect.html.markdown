---
layout: "yandex"
page_title: "Yandex: yandex_mdb_kafka_connect"
sidebar_current: "docs-yandex-mdb-kafka-connect"
description: |-
  Manages a connectors of a Kafka cluster within Yandex.Cloud.
---

# yandex\_mdb\_kafka\_connect

Manages a connector of a Kafka cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kafka/concepts).


## Example Usage

```hcl
resource "yandex_mdb_kafka_cluster" "foo" {
  name        = "foo"
  network_id  = "c64vs98keiqc7f24pvkd"

  config {
    version          = "2.8"
    zones            = ["ru-central1-a"]
    kafka {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-hdd"
        disk_size          = 16
      }
    }
  }
}

resource "yandex_mdb_kafka_connector" connector {
  cluster_id         = yandex_mdb_kafka_cluster.foo.id
  name               = "replication"
  tasks_max          = 3
  properties = {
          refresh.topics.enabled = "true"
  }
  connector_config_mirrormaker {
          topics = "data.*"
          replication_factor = 1
          source_cluster {
                  alias = "source"
                  external_cluster {
                          bootstrap_servers = "somebroker1:9091,somebroker2:9091"
                          sasl_username = "someuser"
                          sasl_password = "somepassword"
                          sasl_mechanism = "SCRAM-SHA-512"
                          security_protocol = "SASL_SSL"
                  }
          }
          target_cluster {
                  alias = "target"
                  this_cluster {}
          }
  }
}
```

## Argument Reference

The following arguments are supported:
* `name` - (Required) The name of the connector.
* `tasks_max` - (Optional) The number of the connector's parallel working tasks. Default is the number of brokers
* `properties` - (Optional) Additional properties for connector.
* `connector_config_mirrormaker` - (Optional) Params for MirrorMaker2 connector. The structure is documented below.

The `connector_config_mirrormaker` block supports:
* `topics` - (Required) The pattern for topic names to be replicated.
* `replication_factor` - (Optional) Replication factor for topics created in target cluster
* `source_cluster` - (Required) Settings for source cluster. The structure is documented below.
* `target_cluster` - (Required) Settings for target cluster. The structure is documented below.

The `source_cluster` and `target_cluster` block supports:
* `alias` - (Optional) Name of the cluster. Used also as a topic prefix
* `external_cluster` - (Optional) Connection params for external cluster
* `this_cluster` - (Optional) Using this section in the cluster definition (source or target) means it's this cluster

The `external_cluster` blocks support:
* `bootstrap_servers` - (Required) List of bootstrap servers to connect to cluster
* `sasl_username` - (Optional) Username to use in SASL authentification mechanism
* `sasl_password` - (Optional) Password to use in SASL authentification mechanism
* `sasl_mechanism` - (Optional) Type of SASL authentification mechanism to use
* `security_protocol` - (Optional) Security protocol to use

## Import

Kafka connector can be imported using following format:

```
$ terraform import yandex_mdb_kafka_connector.foo {{cluster_id}}:{{connector_name}}
```