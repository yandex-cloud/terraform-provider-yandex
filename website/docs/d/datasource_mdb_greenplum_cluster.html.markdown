---
layout: "yandex"
page_title: "Yandex: yandex_mdb_greenplum_cluster"
sidebar_current: "docs-yandex-datasource-mdb-greenplum-cluster"
description: |-
  Get information about a Yandex Managed Greenplum cluster.
---

# yandex\_mdb\_greenplum\_cluster

Get information about a Yandex Managed Greenplum cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-greenplum/).

## Example Usage

```hcl
data "yandex_mdb_greenplum_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_greenplum_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Greenplum cluster.

* `name` - (Optional) The name of the Greenplum cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:


* `network_id` - ID of the network, to which the Greenplum cluster belongs.
* `created_at` - Timestamp of cluster creation.
* `description` - Description of the Greenplum cluster.
* `labels` - A set of key/value label pairs to assign to the Greenplum cluster.
* `environment` - Deployment environment of the Greenplum cluster.
* `version` - Version of the Greenplum cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.

* `zone` - The availability zone where the Greenplum hosts will be created.
* `subnet_id` - The ID of the subnet, to which the hosts belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the master hosts should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.
* `master_host_count` - Number of hosts in master subcluster.
* `segment_host_count` - Number of hosts in segment subcluster.
* `segment_in_host` - Number of segments on segment host.

* `master_subcluster` - Settings for master subcluster. The structure is documented below.
* `segment_subcluster` - Settings for segment subcluster. The structure is documented below.

* `master_hosts` - Info about hosts in master subcluster. The structure is documented below.
* `segment_hosts` - Info about hosts in segment subcluster. The structure is documented below.

* `access` - Access policy to the Greenplum cluster. The structure is documented below.
* `maintenance_window` - Maintenance window settings of the Greenplum cluster. The structure is documented below.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `pooler_config` - Configuration of the connection pooler. The structure is documented below.
* `greenplum_config` - Greenplum cluster config.

* `user_name` - Greenplum cluster admin user name.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `deletion_protection` - Flag to protect the cluster from deletion.

The `master_subcluster` block supports:
* `resources` - Resources allocated to hosts for master subcluster of the Greenplum cluster. The structure is documented below.

The `segment_subcluster` block supports:
* `resources` - Resources allocated to hosts for segment subcluster of the Greenplum cluster. The structure is documented below.

The `master_hosts` block supports:
* `assign_public_ip` - Flag that indicates whether master hosts was created with a public IP.
* `fqdn` - The fully qualified domain name of the host.

The `segment_hosts` block supports:
* `fqdn` - The fully qualified domain name of the host.

The `resources` block supports:
* `resources_preset_id` - The ID of the preset for computational resources available to a Greenplum host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-greenplum/concepts/instance-types).
* `disk_size` - Volume of the storage available to a Greenplum host, in gigabytes.
* `disk_type_id` - Type of the storage for Greenplum hosts.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `data_lens` - Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).
* `web_sql` - Allows access for SQL queries in the management console.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`.
* `day` - Day of the week (in `DDD` format). Value is one of: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"
* `hour` - Hour of the day in UTC (in `HH` format). Value is between 1 and 24.

The `pooler_config` block supports:

* `pooling_mode` - Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.
* `pool_size` - Value for `pool_size` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_size-integer).
* `pool_client_idle_timeout` - Value for `pool_client_idle_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_ttl-integer).

