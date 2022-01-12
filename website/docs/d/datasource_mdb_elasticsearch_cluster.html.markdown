---
layout: "yandex"
page_title: "Yandex: yandex_mdb_elasticsearch_cluster"
sidebar_current: "docs-yandex-datasource-mdb-elasticsearch-cluster"
description: |-
  Get information about a Yandex Managed Elasticsearch cluster.
---

# yandex\_mdb\_elasticsearch\_cluster

Get information about a Yandex Managed Elasticsearch cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-elasticsearch/concepts).

## Example Usage

```hcl
data "yandex_mdb_elasticsearch_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_elasticsearch_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Elasticsearch cluster.

* `name` - (Optional) The name of the Elasticsearch cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the Elasticsearch cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the Elasticsearch cluster.
* `labels` - A set of key/value label pairs to assign to the Elasticsearch cluster.
* `environment` - Deployment environment of the Elasticsearch cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `config` - Configuration of the Elasticsearch cluster. The structure is documented below.
* `host` - A host of the Elasticsearch cluster. The structure is documented below.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `service_account_id` - ID of the service account authorized for this cluster.

The `config` block supports:

* `version` - Version of Elasticsearch.

* `edition` - Edition of Elasticsearch. For more information, see [the official documentation](https://cloud.yandex.com/en-ru/docs/managed-elasticsearch/concepts/es-editions).

* `plugins` - A set of requested Elasticsearch plugins.

* `data_node` - Configuration for Elasticsearch data nodes subcluster. The structure is documented below.

* `master_node` - Configuration for Elasticsearch master nodes subcluster. The structure is documented below.

The `data_node` block supports:

* `resources` - Resources allocated to hosts of the Elasticsearch data nodes subcluster. The structure is documented below.

The `master_node` block supports:

* `resources` - Resources allocated to hosts of the Elasticsearch master nodes subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a Elasticsearch host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/en-ru/docs/managed-elasticsearch/concepts/instance-types).
* `disk_size` - Volume of the storage available to a Elasticsearch host, in gigabytes.
* `disk_type_id` - Type of the storage of Elasticsearch hosts.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `type` - The type of the host to be deployed. For more information, see [the official documentation](https://cloud.yandex.com/en-ru/docs/managed-elasticsearch/concepts/hosts-roles).
* `zone` - The availability zone where the Elasticsearch host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation.

The `maintenance_window` block supports:

* `type` - Type of a maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour need to be specified with the weekly window.
* `hour` - Hour of the day in UTC time zone (1-24) for a maintenance window if the window type is weekly.
* `day` - Day of the week for a maintenance window if the window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.