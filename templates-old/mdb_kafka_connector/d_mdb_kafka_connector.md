---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a connector of the Yandex Managed Kafka cluster.
---

# {{.Name}} ({{.Type}})

Get information about a connector of the Yandex Managed Kafka cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

## Example usage

{{ tffile "examples/mdb_kafka_connector/d_mdb_kafka_connector_1.tf" }}

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kafka cluster.
* `name` - (Required) The name of the Kafka connector.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `partitions` - The number of the topic's partitions.
* `tasks_max` - The number of the connector's parallel working tasks. Default is the number of brokers
* `properties` - Additional properties for connector.
* `connector_config_mirrormaker` - Params for MirrorMaker2 connector. The structure is documented below.
* `connector_config_s3_sink` - Params for S3 Sink connector. The structure is documented below.

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

The `connector_config_s3_sink` block supports:
* `topics` - The pattern for topic names to be copied to s3 bucket.
* `file_compression_type` - Сompression type for messages. Cannot be changed.
* `file_max_records` - Max records per file.
* `s3_connection` - Settings for connection to s3-compatible storage. The structure is documented below.

The `s3_connection` block supports:
* `bucket_name` - Name of the bucket in s3-compatible storage.
* `external_s3` - Connection params for external s3-compatible storage. The structure is documented below.

The `external_s3` blocks support:
* `endpoint` - URL of s3-compatible storage.
* `access_key_id` - ID of aws-compatible static key.
* `secret_access_key` - Secret key of aws-compatible static key.
* `region` - region of s3-compatible storage. [Available region list](https://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/regions/Regions.html).
