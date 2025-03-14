---
subcategory: "Data Processing"
page_title: "Yandex: yandex_dataproc_cluster"
description: |-
  Manages a Data Processing cluster within Yandex Cloud.
---

# yandex_dataproc_cluster (Resource)

Manages a Yandex Data Processing cluster. For more information, see [the official documentation](https://yandex.cloud/docs/data-proc/).

## Example usage

```terraform
//
// Create a new Data Processing Cluster.
//
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
    # Certain cluster version can be set, but better to use default value (last stable version)
    # version_id = "2.0"

    hadoop {
      services = ["HDFS", "YARN", "SPARK", "TEZ", "MAPREDUCE", "HIVE"]
      properties = {
        "yarn:yarn.resourcemanager.am.max-attempts" = 5
      }
      ssh_public_keys = [
      file("~/.ssh/id_rsa.pub")]
      initialization_action {
        uri  = "s3a://yandex_storage_bucket.foo.bucket/scripts/script.sh"
        args = ["arg1", "arg2"]
      }
    }

    subcluster_spec {
      name = "main"
      role = "MASTERNODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
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
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
    }

    subcluster_spec {
      name = "compute"
      role = "COMPUTENODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
    }

    subcluster_spec {
      name = "compute_autoscaling"
      role = "COMPUTENODE"
      resources {
        resource_preset_id = "s2.small"
        disk_type_id       = "network-hdd"
        disk_size          = 20
      }
      subnet_id   = yandex_vpc_subnet.foo.id
      hosts_count = 2
      autoscaling_config {
        max_hosts_count        = 10
        measurement_duration   = 60
        warmup_duration        = 60
        stabilization_duration = 120
        preemptible            = false
        decommission_timeout   = 60
      }
    }
  }
}

// Auxiliary resources for Data Processing Cluster
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

* `name` - (Required) Name of a specific Yandex Data Processing cluster.
* `cluster_config` - (Required) Configuration and resources for hosts that should be created with the cluster. The structure is documented below.
* `service_account_id` - (Required) Service account to be used by the Yandex Data Processing agent to access resources of Yandex Cloud. Selected service account should have `mdb.dataproc.agent` role on the folder where the Yandex Data Processing cluster will be located.

---

* `folder_id` - (Optional) ID of the folder to create a cluster in. If it is not provided, the default provider folder is used.
* `bucket` - (Optional) Name of the Object Storage bucket to use for Yandex Data Processing jobs. Yandex Data Processing Agent saves output of job driver's process to specified bucket. In order for this to work service account (specified by the `service_account_id` argument) should be given permission to create objects within this bucket.
* `description` - (Optional) Description of the Yandex Data Processing cluster.
* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Data Processing cluster.
* `zone_id` - (Optional) ID of the availability zone to create cluster in. If it is not provided, the default provider zone is used.
* `ui_proxy` - (Optional) Whether to enable UI Proxy feature.
* `security_group_ids` - (Optional) A list of security group IDs that the cluster belongs to.
* `host_group_ids` - (Optional) A list of host group IDs to place VMs of the cluster on.
* `deletion_protection` - (Optional) Inhibits deletion of the cluster. Can be either `true` or `false`.
* `log_group_id` - (Optional) ID of the cloud logging group for cluster logs.
* `environment` - (Optional) Deployment environment of the cluster. Can be either `PRESTABLE` or `PRODUCTION`. The default is `PRESTABLE`.

---

The `cluster_config` block supports:

* `version_id` - (Required) Version of Yandex Data Processing image.
* `hadoop` - (Optional) Yandex Data Processing specific options. The structure is documented below.
* `subcluster_spec` - (Required) Configuration of the Yandex Data Processing subcluster. The structure is documented below.

---

The `hadoop` block supports:

* `services` - (Optional) List of services to run on Yandex Data Processing cluster.
* `properties` - (Optional) A set of key/value pairs that are used to configure cluster services.
* `ssh_public_keys` - (Optional) List of SSH public keys to put to the hosts of the cluster. For information on how to connect to the cluster, see [the official documentation](https://yandex.cloud/docs/data-proc/operations/connect).
* `initialization_action` - (Optional) List of initialization scripts. The structure is documented below.

---

The `initialization_action` block supports:

* `uri` - (Required) Script URI.
* `args` - (Optional) List of arguments of the initialization script.
* `timeout` - (Optional) Script execution timeout, in seconds.

---

The `subcluster_spec` block supports:

* `name` - (Required) Name of the Yandex Data Processing subcluster.
* `role` - (Required) Role of the subcluster in the Yandex Data Processing cluster.
* `resources` - (Required) Resources allocated to each host of the Yandex Data Processing subcluster. The structure is documented below.
* `subnet_id` - (Required) The ID of the subnet, to which hosts of the subcluster belong. Subnets of all the subclusters must belong to the same VPC network.
* `hosts_count` - (Required) Number of hosts within Yandex Data Processing subcluster.
* `assign_public_ip` - (Optional) If true then assign public IP addresses to the hosts of the subclusters.
* `autoscaling_config` - (Optional) Autoscaling configuration for compute subclusters.

---

The `resources` block supports:

* `resource_preset_id` - (Required) The ID of the preset for computational resources available to a host. All available presets are listed in the [documentation](https://yandex.cloud/docs/data-proc/concepts/instance-types).
* `disk_size` - (Required) Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - (Optional) Type of the storage of a host. One of `network-hdd` (default) or `network-ssd`.

---

The `autoscaling_config` block supports:

* `max_hosts_count` - (Required) Maximum number of nodes in autoscaling subclusters.
* `preemptible` - (Optional) Bool flag -- whether to use preemptible compute instances. Preemptible instances are stopped at least once every 24 hours, and can be stopped at any time if their resources are needed by Compute. For more information, see [Preemptible Virtual Machines](https://yandex.cloud/docs/compute/concepts/preemptible-vm).
* `warmup_duration` - (Optional) The warmup time of the instance in seconds. During this time, traffic is sent to the instance, but instance metrics are not collected.
* `stabilization_duration` - (Optional) Minimum amount of time in seconds allotted for monitoring before Instance Groups can reduce the number of instances in the group. During this time, the group size doesn't decrease, even if the new metric values indicate that it should.
* `measurement_duration` - (Optional) Time in seconds allotted for averaging metrics.
* `cpu_utilization_target` - (Optional) Defines an autoscaling rule based on the average CPU utilization of the instance group. If not set default autoscaling metric will be used.
* `decommission_timeout` - (Optional) Timeout to gracefully decommission nodes during downscaling. In seconds.

## Attributes Reference

* `id` - (Computed) ID of a new Yandex Data Processing cluster.
* `created_at` - (Computed) The Yandex Data Processing cluster creation timestamp.
* `cluster_config.0.subcluster_spec.X.id` - (Computed) ID of the subcluster.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_dataproc_cluster.<resource Name> <resource Id> 
terraform import yandex_dataproc_cluster.foo c9q85**********gpn21
```
