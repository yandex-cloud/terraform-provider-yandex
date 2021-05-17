---
layout: "yandex"
page_title: "Yandex: yandex_dataproc_cluster"
sidebar_current: "docs-yandex-dataproc-cluster"
description: |-
  Manages a Data Proc cluster within Yandex.Cloud.
---

# yandex\_dataproc\_cluster

Manages a Data Proc cluster. For more information, see [the official documentation](https://cloud.yandex.com/docs/data-proc/).

## Example Usage

```hcl
resource "yandex_dataproc_cluster" "foo" {
  depends_on = [yandex_resourcemanager_folder_iam_binding.dataproc]

  bucket      = yandex_storage_bucket.foo.bucket
  description = "Dataproc Cluster created by Terraform"
  name        = "dataproc-cluster"
  labels = {
    created_by = "terraform"
  }
  service_account_id = yandex_iam_service_account.dataproc.id
  zone_id            = "ru-central1-b"

  cluster_config {
    version_id = "1.0"

    hadoop {
      services = ["HDFS", "YARN", "SPARK", "TEZ", "MAPREDUCE", "HIVE"]
      properties = {
        "yarn:yarn.resourcemanager.am.max-attempts" = 5
      }
      ssh_public_keys = [
      file("~/.ssh/id_rsa.pub")]
    }

    subcluster_spec {
      name = "main"
      role = "MASTERNODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 16
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 1
    }

    subcluster_spec {
      name = "data"
      role = "DATANODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 16
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
    }
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_iam_service_account" "dataproc" {
  name        = "dataproc"
  description = "service account to manage Dataproc Cluster"
}

data "yandex_resourcemanager_folder" "foo" {
  folder_id = "some_folder_id"
}

resource "yandex_resourcemanager_folder_iam_binding" "dataproc" {
  folder_id = data.yandex_resourcemanager_folder.foo.id
  role      = "mdb.dataproc.agent"
  members = [
    "serviceAccount:${yandex_iam_service_account.dataproc.id}",
  ]
}

// required in order to create bucket
resource "yandex_resourcemanager_folder_iam_binding" "bucket-creator" {
  folder_id = data.yandex_resourcemanager_folder.foo.id
  role      = "editor"
  members = [
    "serviceAccount:${yandex_iam_service_account.dataproc.id}",
  ]
}

resource "yandex_iam_service_account_static_access_key" "foo" {
  service_account_id = yandex_iam_service_account.dataproc.id
}

resource "yandex_storage_bucket" "foo" {
  depends_on = [
    yandex_resourcemanager_folder_iam_binding.bucket-creator
  ]

  bucket     = "foo"
  access_key = yandex_iam_service_account_static_access_key.foo.access_key
  secret_key = yandex_iam_service_account_static_access_key.foo.secret_key
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of a specific Data Proc cluster.
* `cluster_config` - (Required) Configuration and resources for hosts that should be created with the cluster. The structure is documented below.
* `service_account_id` - (Required) Service account to be used by the Data Proc agent to access resources of Yandex.Cloud. Selected service account should have `mdb.dataproc.agent` role on the folder where the Data Proc cluster will be located.
* `ui_proxy` - (Optional) Whether to enable UI Proxy feature.
* `security_group_ids` - (Optional) Security group ids, to which cluster belongs.

---

* `folder_id` - (Optional) ID of the folder to create a cluster in. If it is not provided, the default provider folder is used.
* `bucket` - (Optional) Name of the Object Storage bucket to use for Data Proc jobs. Data Proc Agent saves output of job driver's process to specified bucket. In order for this to work service account (specified by the `service_account_id` argument) should be given permission to create objects within this bucket.
* `description` - (Optional) Description of the Data Proc cluster.
* `labels` - (Optional) A set of key/value label pairs to assign to the Data Proc cluster.
* `zone_id` - (Optional) ID of the availability zone to create cluster in. If it is not provided, the default provider zone is used.

---

The `cluster_config` block supports:

* `version_id` - (Required) Version of Data Proc image.
* `hadoop` - (Optional) Data Proc specific options. The structure is documented below.
* `subcluster_spec` - (Required) Configuration of the Data Proc subcluster. The structure is documented below.

---

The `hadoop` block supports:

* `services` - (Optional) List of services to run on Data Proc cluster.
* `properties` - (Optional) A set of key/value pairs that are used to configure cluster services.
* `ssh_public_keys` - (Optional) List of SSH public keys to put to the hosts of the cluster. For information on how to connect to the cluster, see [the official documentation](https://cloud.yandex.com/docs/data-proc/operations/connect).

---

The `subcluster_spec` block supports:

* `name` - (Required) Name of the Data Proc subcluster.
* `role` - (Required) Role of the subcluster in the Data Proc cluster.
* `resources` - (Required) Resources allocated to each host of the Data Proc subcluster. The structure is documented below.
* `subnet_id` - (Required) The ID of the subnet, to which hosts of the subcluster belong. Subnets of all the subclusters must belong to the same VPC network.
* `hosts_count` - (Required) Number of hosts within Data Proc subcluster.

---

The `resources` block supports:

* `resource_preset_id` - (Required) The ID of the preset for computational resources available to a host. All available presets are listed in the [documentation](https://cloud.yandex.com/docs/data-proc/concepts/instance-types).
* `disk_size` - (Required) Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - (Optional) Type of the storage of a host. One of `network-hdd` (default) or `network-ssd`.

## Attributes Reference

* `id` - (Computed) ID of a new Data Proc cluster.
* `created_at` - (Computed) The Data Proc cluster creation timestamp.
* `cluster_config.0.subcluster_spec.X.id` - (Computed) ID of the subcluster.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_dataproc_cluster.foo cluster_id
```
