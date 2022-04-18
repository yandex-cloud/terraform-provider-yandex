---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mongodb_cluster"
sidebar_current: "docs-yandex-datasource-mdb-mongodb-cluster"
description: |-
  Get information about a Yandex Managed MongoDB cluster.
---

# yandex\_mdb\_mongodb\_cluster

Get information about a Yandex Managed MongoDB cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts).

## Example Usage

```hcl
data "yandex_mdb_mongodb_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_mongodb_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the MongoDB cluster.
* `name` - (Optional) The name of the MongoDB cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `description` - Description of the MongoDB cluster.
* `network_id` - ID of the network, to which the MongoDB cluster belongs.
* `environment` - Deployment environment of the MongoDB cluster.
* `created_at` - Creation timestamp of the key.
* `labels` - A set of key/value label pairs to assign to the MongoDB cluster.
* `sharded` - MongoDB Cluster mode enabled/disabled.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `resources` - Resources allocated to hosts of the MongoDB cluster. The structure is documented below.
* `host` - A host of the MongoDB cluster. The structure is documented below.
* `cluster_config` - Configuration of the MongoDB cluster. The structure is documented below.
* `user` - A user of the MongoDB cluster. The structure is documented below.
* `database` - A database of the MongoDB cluster. The structure is documented below.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts/instance-types).
* `disk_size` - Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - The ID of the storage type. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts/storage)

The `host` block supports:

* `name` - The fully qualified domain name of the host.
* `zone_id` - The availability zone where the MongoDB host will be created.
* `role` - The role of the cluster (either PRIMARY or SECONDARY).
* `health` - The health of the host.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must
  be a part of the network to which the cluster belongs.
* `assign_public_ip` - Has assigned public IP.
* `shard_name` - The name of the shard to which the host belongs.
* `type` - type of mongo demon which runs on this host (mongod, mongos or monogcfg).

The `cluster_config` block supports:

* `version` - Version of MongoDB (either 5.0, 5.0-enterprise, 4.4, 4.4-enterprise, 4.2, 4.0 or 3.6).
* `feature_compatibility_version` - Feature compatibility version of MongoDB.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.
* `access` - Access policy to MongoDB cluster. The structure is documented below.
* `mongod` - (Optional) Configuration of the mongod service. The structure is documented below.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `data_lens` - Shows whether cluster has access to data lens.

The `user` block supports:

* `name` - The name of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.
* `roles` - (Optional) List of strings. The roles of the user in this database. For more information see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts/users-and-roles).

The `database` block supports:

* `name` - The name of the database.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.


The `mongod` block supports:

* `security` - (Optional) A set of MongoDB Security settings
  (see the [security](https://www.mongodb.com/docs/manual/reference/configuration-options/#security-options) option).
  The structure is documented below. Available only in enterprise edition.

* `audit_log` - (Optional) A set of audit log settings
  (see the [auditLog](https://www.mongodb.com/docs/manual/reference/configuration-options/#auditlog-options) option).
  The structure is documented below. Available only in enterprise edition.

* `set_parameter` - (Optional) A set of MongoDB Server Parameters
  (see the [setParameter](https://www.mongodb.com/docs/manual/reference/configuration-options/#setparameter-option) option).
  The structure is documented below.


The `security` block supports:

* `enable_encryption` - (Optional) Enables the encryption for the WiredTiger storage engine. Can be either true or false.
  For more information see [security.enableEncryption](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.enableEncryption)
  description in the official documentation. Available only in enterprise edition.

* `kmip` - (Optional) Configuration of the third party key management appliance via the Key Management Interoperability Protocol (KMIP)
  (see [Encryption tutorial](https://www.mongodb.com/docs/rapid/tutorial/configure-encryption) ). Requires `enable_encryption` to be true.
  The structure is documented below. Available only in enterprise edition.


The `audit_log` block supports:

* `filter` - (Optional) Configuration of the audit log filter in JSON format.
  For more information see [auditLog.filter](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-auditLog.filter)
  description in the official documentation. Available only in enterprise edition.


The `set_parameter` block supports:

* `audit_authorization_success` - (Optional) Enables the auditing of authorization successes. Can be either true or false.
  For more information, see the [auditAuthorizationSuccess](https://www.mongodb.com/docs/manual/reference/parameters/#mongodb-parameter-param.auditAuthorizationSuccess)
  description in the official documentation. Available only in enterprise edition.


The `kmip` block supports:

* `server_name` - (Required) Hostname or IP address of the KMIP server to connect to.
  For more information see [security.kmip.serverName](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.serverName)
  description in the official documentation.

* `port` - (Optional) Port number to use to communicate with the KMIP server. Default: 5696
  For more information see [security.kmip.port](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.port)
  description in the official documentation.

* `server_ca` - (Required) Path to CA File. Used for validating secure client connection to KMIP server.
  For more information see [security.kmip.serverCAFile](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.serverCAFile)
  description in the official documentation.

* `client_certificate` - (Required) String containing the client certificate used for authenticating MongoDB to the KMIP server.
  For more information see [security.kmip.clientCertificateFile](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.clientCertificateFile)
  description in the official documentation.

* `key_identifier` - (Optional) Unique KMIP identifier for an existing key within the KMIP server.
  For more information see [security.kmip.keyIdentifier](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.keyIdentifier)
  description in the official documentation.
