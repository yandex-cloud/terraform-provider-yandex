---
layout: "yandex"
page_title: "Yandex: yandex_mdb_greenplum_cluster"
sidebar_current: "docs-yandex-mdb-greenplum-cluster"
description: |-
  Manages a Greenplum cluster within Yandex.Cloud.
---

# yandex\_mdb\_greenplum\_cluster

Manages a Greenplum cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.ru/docs/managed-greenplum/).

Please read [Pricing for Managed Service for Greenplum](https://cloud.yandex.ru/docs/managed-greenplum/) before using Greenplum cluster.

Yandex Managed Service for Greenplum® is now in preview

## Example Usage

Example of creating a Single Node Greenplum.

```hcl
resource "yandex_mdb_greenplum_cluster" "foo" {
  name               = "test"
  description        = "test greenplum cluster"
  environment        = "PRESTABLE"
  network_id         = yandex_vpc_network.foo.id
  zone_id            = "ru-central1-a"
  subnet_id          = yandex_vpc_subnet.foo.id
  assign_public_ip   = true
  version            = "6.19"
  master_host_count  = 2
  segment_host_count = 5
  segment_in_host    = 1
  master_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }
  segment_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }

  access {
    web_sql = true
  }

  greenplum_config = {
    max_connections                   = 395
    gp_workfile_compression           = "false"
  }

  user_name     = "admin_user"
  user_password = "your_super_secret_password"

  security_group_ids = [yandex_vpc_security_group.test-sg-x.id]
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}

resource "yandex_vpc_security_group" "test-sg-x" {
  network_id = yandex_vpc_network.foo.id
  ingress {
    protocol       = "ANY"
    description    = "Allow incoming traffic from members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol       = "ANY"
    description    = "Allow outgoing traffic to members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Greenplum cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the Greenplum cluster uses.

* `zone` - (Required) The availability zone where the Greenplum hosts will be created.

* `subnet_id` - (Required) The ID of the subnet, to which the hosts belongs. The subnet must be a part of the network to which the cluster belongs.

* `assign_public_ip` - (Required) Sets whether the master hosts should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.


* `environment` - (Required) Deployment environment of the Greenplum cluster. (PRODUCTION, PRESTABLE)

* `version` - (Required) Version of the Greenplum cluster. (6.19)


* `master_host_count` - (Required) Number of hosts in master subcluster (1 or 2).

* `segment_host_count` - (Required) Number of hosts in segment subcluster (from 1 to 32).

* `segment_in_host` - (Required) Number of segments on segment host (not more then 1 + RAM/8).

* `master_subcluster` - (Required) Settings for master subcluster. The structure is documented below.

* `segment_subcluster` - (Required) Settings for segment subcluster. The structure is documented below.

* `access` - (Optional) Access policy to the Greenplum cluster. The structure is documented below.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `pooler_config` - (Optional) Configuration of the connection pooler. The structure is documented below.

* `greenplum_config` - (Optional) Greenplum cluster config. Detail info in "Greenplum cluster settings" section (documented below).

- - -
* `user_name` - (Required) Greenplum cluster admin user name.

* `user_password` - (Required) Greenplum cluster admin password name.

- - -

* `description` - (Optional) Description of the Greenplum cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Greenplum cluster.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.

- - -

The `master_subcluster` block supports:
* `resources` - (Required) Resources allocated to hosts for master subcluster of the Greenplum cluster. The structure is documented below.

The `segment_subcluster` block supports:
* `resources` - (Required) Resources allocated to hosts for segment subcluster of the Greenplum cluster. The structure is documented below.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started (UTC).

* `minutes` - (Optional) The minute at which backup will be started (UTC).

The `access` block supports:

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).

* `web_sql` - Allows access for SQL queries in the management console

The `pooler_config` block supports:

* `pooling_mode` - (Optional) Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.

* `pool_size` - (Optional) Value for `pool_size` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_size-integer).

* `pool_client_idle_timeout` - (Optional) Value for `pool_client_idle_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_ttl-integer).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the cluster.

* `health` - Aggregated health of the cluster.

* `status` - Status of the cluster.

- - -
* `master_hosts` - (Computed) Info about hosts in master subcluster. The structure is documented below.

* `segment_hosts` - (Computed) Info about hosts in segment subcluster. The structure is documented below.

- - -
The `master_hosts` block supports:
* `assign_public_ip` - (Computed) Flag indicating that master hosts should be created with a public IP address.
* `fqdn` - (Computed) The fully qualified domain name of the host.

The `segment_hosts` block supports:
* `fqdn` - (Computed) The fully qualified domain name of the host.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_greenplum_cluster.foo cluster_id
```

## Greenplum cluster settings

| Setting name and type \ Greenplum version | 6.17 | 6.19 |
| ------------------------------------------| ---- | ---- |
| max_connections : integer | supported | supported |
| max_prepared_transactions : integer | supported | supported |
| gp_workfile_limit_per_query : integer | supported | supported |
| gp_workfile_limit_files_per_query : integer | supported | supported |
| max_slot_wal_keep_size : integer | supported | supported |
| gp_workfile_limit_per_segment : integer | supported | supported |
| gp_workfile_compression : boolean | supported | supported |
| max_statement_mem : integer | - | supported |
| log_statement : one of<br>  - 0: " LOG_STATEMENT_UNSPECIFIED"<br>  - 1: " LOG_STATEMENT_NONE"<br>  - 2: " LOG_STATEMENT_DDL"<er>  - 3: " LOG_STATEMENT_MOD"<br>  - 4: " LOG_STATEMENT_ALL"  | - | supported |