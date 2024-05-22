---
layout: "yandex"
page_title: "Yandex: yandex_mdb_opensearch_cluster"
sidebar_current: "docs-yandex-mdb-opensearch-cluster"
description: |-
  Manages a OpenSearch cluster within Yandex.Cloud.
---

# yandex\_mdb\_opensearch\_cluster

Manages a OpenSearch cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-opensearch/concepts).

## Example Usage

Example of creating a Single Node OpenSearch cluster.

```hcl
resource "yandex_mdb_opensearch_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  config {

    admin_password = "super-password"

    opensearch {
      node_groups {
          name = "group0"
          assign_public_ip     = true
          hosts_count          = 1
          subnet_ids           = ["${yandex_vpc_subnet.foo.id}"]
          zone_ids             = ["ru-central1-a"]
          roles                = ["data", "manager"]
          resources {
            resource_preset_id   = "s2.micro"
            disk_size            = 10737418240
            disk_type_id         = "network-ssd"
          }
      }
    }
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

Example of creating a high available OpenSearch Cluster.

```hcl
locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-c",
  ]
}

resource "yandex_mdb_opensearch_cluster" "foo" {
  name        = "my-cluster"
  environment = "PRODUCTION"
  network_id  = "${yandex_vpc_network.es-net.id}"

  config {

    admin_password = "super-password"

    opensearch {
      node_groups {
          name = "hot_group0"
          assign_public_ip     = true
          hosts_count          = 2
          zone_ids             = local.zones
          roles                = ["data"]
          resources {
            resource_preset_id   = "s2.small"
            disk_size            = 10737418240
            disk_type_id         = "network-ssd"
          }
      }

      node_groups {
          name = "cold_group0"
          assign_public_ip     = true
          hosts_count          = 2
          zone_ids             = local.zones
          roles                = ["data"]
          resources {
            resource_preset_id   = "s2.micro"
            disk_size            = 10737418240
            disk_type_id         = "network-hdd"
          }
      }

      node_groups {
          name = "managers_group"
          assign_public_ip     = true
          hosts_count          = 3
          zone_ids             = local.zones
          roles                = ["manager"]
          resources {
            resource_preset_id   = "s2.micro"
            disk_size            = 10737418240
            disk_type_id         = "network-ssd"
          }
      }

      plugins = ["analysis-icu"]
    }

    dashboards {
      node_groups {
          name = "dashboards"
          assign_public_ip     = true
          hosts_count          = 1
          zone_ids             = local.zones
          resources {
            resource_preset_id   = "s2.micro"
            disk_size            = 10737418240
            disk_type_id         = "network-ssd"
          }
      }
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

* `name` - (Required) Name of the OpenSearch cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the OpenSearch cluster belongs.

* `config` - (Required) Configuration of the OpenSearch cluster. The structure is documented below.

- - -

* `environment` - (Optional) Deployment environment of the OpenSearch cluster. Can be either `PRESTABLE` or `PRODUCTION`. Default: `PRODUCTION`

* `description` - (Optional) Description of the OpenSearch cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the OpenSearch cluster.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `service_account_id` - (Optional) ID of the service account authorized for this cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.

- - -

The `config` block supports:

* `version` - (Optional) Version of OpenSearch.

* `admin_password` - (Required) Password for admin user of OpenSearch.

* `opensearch` - (Required) Configuration for OpenSearch node groups. The structure is documented below.

* `dashboards` - (Optional) Configuration for Dashboards node groups. The structure is documented below.

The `opensearch` block supports:

* `plugins` - (Optional) A set of requested OpenSearch plugins.

* `node_groups` - (Required) A set of named OpenSearch node group configurations. The structure is documented below.

The OpenSearch `node_groups` block supports:

* `name` - (Required) Name of OpenSearch node group.

* `resources` - (Required) Resources allocated to hosts of this OpenSearch node group. The structure is documented below.

* `host_count` - (Required) Number of hosts in this node group.

* `zones_ids` - (Required) A set of availability zones where hosts of node group may be allocated. No other parameters should be changed simultaneously with this one, except `subnet_ids`.

* `subnet_ids` - (Optional) A set of the subnets, to which the hosts belongs. The subnets must be a part of the network to which the cluster belongs. No other parameters should be changed simultaneously with this one, except `zones_ids`.

* `assign_public_ip` - (Optional) Sets whether the hosts should get a public IP address on creation.

* `roles` - (Optional) A set of OpenSearch roles assigned to hosts. Available roles are: `DATA`, `MANAGER`. Default: [`DATA`, `MANAGER`]

The Dashboards `node_groups` block supports:

* `name` - (Required) Name of OpenSearch node group.

* `resources` - (Required) Resources allocated to hosts of this Dashboards node group. The structure is documented below.

* `host_count` - (Required) Number of hosts in this node group.

* `zones_ids` - (Required) A set of availability zones where hosts of node group may be allocated. No other parameters should be changed simultaneously with this one, except `subnet_ids`.

* `subnet_ids` - (Optional) A set of the subnets, to which the hosts belongs. The subnets must be a part of the network to which the cluster belongs. No other parameters should be changed simultaneously with this one, except `zones_ids`.

* `assign_public_ip` - (Optional) Sets whether the hosts should get a public IP address on creation.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-opensearch/concepts).

* `disk_size` - (Required) Volume of the storage available to a host, in bytes.

* `disk_type_id` - (Required) Type of the storage of OpenSearch hosts.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-opensearch/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-opensearch/api-ref/Cluster/).

* `hosts` - A hosts of the OpenSearch cluster. The structure is documented below.

The `hosts` block supports:

* `fqdn` - The fully qualified domain name of the host.

* `zone` - The availability zone where the OpenSearch host will be created.
  For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).

* `type` - The type of the deployed host. Can be either `OPENSEARCH` or `DASHBOARDS`.

* `roles` - The roles of the deployed host. Can contain `DATA` and/or `MANAGER` roles. Will be empty for `DASHBOARDS` type.

* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must
  be a part of the network to which the cluster belongs.

* `assign_public_ip` - Sets whether the host should get a public IP address. Can be either `true` or `false`.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_opensearch_cluster.foo cluster_id
```
