---
layout: "yandex"
page_title: "Yandex: yandex_mdb_elasticsearch_cluster"
sidebar_current: "docs-yandex-mdb-elasticsearch-cluster"
description: |-
  Manages a Elasticsearch cluster within Yandex.Cloud.
---

# yandex\_mdb\_elasticsearch\_cluster

Manages a Elasticsearch cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-elasticsearch/concepts).

## Example Usage

Example of creating a Single Node Elasticsearch.

```hcl
resource "yandex_mdb_elasticsearch_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  config {

    admin_password = "super-password"

    data_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 100
      }
    }

  }

  host {
    name = "node"
    zone = "ru-central1-a"
    type = "DATA_NODE"
    assign_public_ip = true
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a high available Elasticsearch Cluster.

```hcl
locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-c",
  ]
}

resource "yandex_mdb_elasticsearch_cluster" "foo" {
  name        = "my-cluster"
  environment = "PRODUCTION"
  network_id  = "${yandex_vpc_network.es-net.id}"

  config {

    edition = "platinum"

    admin_password = "super-password"

    data_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 100
      }
    }

    master_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 10
      }
    }

    plugins = ["analysis-icu"]

  }

  dynamic "host" {
    for_each = toset(range(0,6))
    content {
      name = "datanode${host.value}"
      zone = local.zones[(host.value)%3]
      type = "DATA_NODE"
      assign_public_ip = true
    }
  }

  dynamic "host" {
    for_each = toset(range(0,3))
    content {
      name = "masternode${host.value}"
      zone = local.zones[host.value%3]
      type = "MASTER_NODE"
    }
  }

  depends_on = [
    yandex_vpc_subnet.es-subnet-a,
    yandex_vpc_subnet.es-subnet-b,
    yandex_vpc_subnet.es-subnet-c,
  ]

}

resource "yandex_vpc_network" "es-net" {}

resource "yandex_vpc_subnet" "es-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.es-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "es-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.es-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "es-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.es-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Elasticsearch cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the Elasticsearch cluster belongs.

* `environment` - (Required) Deployment environment of the Elasticsearch cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `config` - (Required) Configuration of the Elasticsearch cluster. The structure is documented below.

* `host` - (Required) A host of the Elasticsearch cluster. The structure is documented below.

- - -

* `description` - (Optional) Description of the Elasticsearch cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Elasticsearch cluster.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `service_account_id` - (Optional) ID of the service account authorized for this cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.

- - -

The `config` block supports:

* `version` - (Optional) Version of Elasticsearch.

* `edition` - (Optional) Edition of Elasticsearch. For more information, see [the official documentation](https://cloud.yandex.com/en-ru/docs/managed-elasticsearch/concepts/es-editions).

* `plugins` - (Optional) A set of Elasticsearch plugins to install.

* `admin_password` - (Required) Password for admin user of Elasticsearch.

* `data_node` - (Required) Configuration for Elasticsearch data nodes subcluster. The structure is documented below.

* `master_node` - (Optional) Configuration for Elasticsearch master nodes subcluster. The structure is documented below.

The `data_node` block supports:

* `resources` - (Required) Resources allocated to hosts of the Elasticsearch data nodes subcluster. The structure is documented below.

The `master_node` block supports:

* `resources` - (Required) Resources allocated to hosts of the Elasticsearch master nodes subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-elasticsearch/concepts).

* `disk_size` - (Required) Volume of the storage available to a host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of Elasticsearch hosts.

The `host` block supports:

* `name` (Required) - User defined host name.

* `fqdn` (Computed) - The fully qualified domain name of the host.

* `zone` - (Required) The availability zone where the Elasticsearch host will be created.
  For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).

* `type` - (Required) The type of the host to be deployed. Can be either `DATA_NODE` or `MASTER_NODE`.

* `subnet_id` (Optional) - The ID of the subnet, to which the host belongs. The subnet must
  be a part of the network to which the cluster belongs.

* `assign_public_ip` (Optional) - Sets whether the host should get a public IP address on creation. Can be either `true` or `false`.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-elasticsearch/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-elasticsearch/api-ref/Cluster/).

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_elasticsearch_cluster.foo cluster_id
```
