---
layout: "yandex"
page_title: "Yandex: yandex_kubernetes_cluster"
sidebar_current: "docs-yandex-kubernetes-cluster"
description: |-
  Allows management of Yandex Kubernetes Cluster. For more information, see
  [the official documentation](https://cloud.yandex.com/docs/managed-kubernetes/concepts/#kubernetes-cluster).
---

# yandex\_kubernetes\_cluster

Creates a Yandex Kubernetes Cluster.

## Example Usage

```hcl
resource "yandex_kubernetes_cluster" "zonal_cluster_resource_name" {
  name        = "name"
  description = "description"

  network_id = "${yandex_vpc_network.network_resource_name.id}"

  master {
    zonal {
      zone      = "${yandex_vpc_subnet.subnet_resource_name.zone}"
      subnet_id = "${yandex_vpc_subnet.subnet_resource_name.id}"
    }

    public_ip = true
  }

  service_account_id      = "${yandex_iam_service_account.service_account_resource_name.id}"
  node_service_account_id = "${yandex_iam_service_account.node_service_account_resource_name.id}"

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

  release_channel = "STABLE"
}
```

```hcl
resource "yandex_kubernetes_cluster" "regional_cluster_resource_name" {
  name        = "name"
  description = "description"

  network_id = "${yandex_vpc_network.network_resource_name.id}"

  master {
    regional {
      region = "ru-central1"

      location {
        zone      = "${yandex_vpc_subnet.subnet_a_resource_name.zone}"
        subnet_id = "${yandex_vpc_subnet.subnet_a_resource_name.id}"
      }

      location {
        zone      = "${yandex_vpc_subnet.subnet_b_resource_name.zone}"
        subnet_id = "${yandex_vpc_subnet.subnet_b_resource_name.id}"
      }

      location {
        zone      = "${yandex_vpc_subnet.subnet_c_resource_name.zone}"
        subnet_id = "${yandex_vpc_subnet.subnet_c_resource_name.id}"
      }
    }

    public_ip = true
  }

  service_account_id      = "${yandex_iam_service_account.service_account_resource_name.id}"
  node_service_account_id = "${yandex_iam_service_account.node_service_account_resource_name.id}"

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

  release_channel = "STABLE"
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of a specific Kubernetes cluster.
* `description` - (Optional) A description of the Kubernetes cluster.
* `folder_id` - (Optional) The ID of the folder that the Kubernetes cluster belongs to.
If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Kubernetes cluster.
* `network_id` - (Optional) The ID of the cluster network.

* `cluster_ipv4_range` - (Optional) CIDR block. IP range for allocating pod addresses.
It should not overlap with any subnet in the network the Kubernetes cluster located in. Static routes will be
set up for this CIDR blocks in node subnets.

* `service_ipv4_range` - (Optional) CIDR block. IP range Kubernetes service Kubernetes cluster
IP addresses will be allocated from. It should not overlap with any subnet in the network
the Kubernetes cluster located in.

* `service_account_id` - Service account to be used for provisioning Compute Cloud and VPC resources
for Kubernetes cluster. Selected service account should have `edit` role on the folder where the Kubernetes
cluster will be located and on the folder where selected network resides.

* `node_service_account_id` - Service account to be used by the worker nodes of the Kubernetes cluster
to access Container Registry or to push node logs and metrics.

* `release_channel` - Cluster release channel.

* `master` - IP allocation policy of the Kubernetes cluster.

The structure is documented below.

## Attributes Reference

* `cluster_id` - (Computed) ID of a new Kubernetes cluster.
* `created_at` - The Kubernetes cluster creation timestamp.
* `status` - Status of the Kubernetes cluster.
* `health` - Health of the Kubernetes cluster.

---

The `master` block supports:

* `version` - (Optional) Version of Kubernetes that will be used for master.
* `public_ip` - (Optional) Boolean flag. When `true`, Kubernetes master will have visible ipv4 address. 

* `zonal` - (Optional) Initialize parameters for Zonal Master (one node master).

The structure is documented below.

* `regional` - (Optional) Initialize parameters for Zonal Master (one node master).

The structure is documented below.

* `version_info` - (Computed) Information about cluster version.

The structure is documented below.

* `internal_v4_address` - (Computed) An IPv4 internal network address that is assigned to the master.
* `external_v4_address` - (Computed) An IPv4 external network address that is assigned to the master.
* `internal_v4_endpoint` - (Computed) Internal endpoint that can be used to connect to the master from cloud networks. 
* `external_v4_endpoint` - (Computed) External endpoint that can be used to access Kubernetes cluster API from the internet (outside of the cloud).
* `cluster_ca_certificate` - (Computed) PEM-encoded public certificate that is the root of trust for the Kubernetes cluster.  

---

The `zonal` block supports:

* `zone` - (Optional) ID of the availability zone. 
* `subnet_id` - (Optional) ID of the subnet. If no ID is specified, and there only one subnet in specified zone, an address in this subnet will be allocated.

---

The `regional` block supports:

* `location` - Array of locations, where master will be allocated. 

The structure is documented below.

---

The `location` block supports repeated values:

* `zone` - (Optional) ID of the availability zone. 
* `subnet_id` - (Optional) ID of the subnet.

---

The `version_info` block supports:

* `current_version` - Current Kubernetes version, major.minor (e.g. 1.15).
* `new_revision_available` - Boolean flag.
Newer revisions may include Kubernetes patches (e.g 1.15.1 -> 1.15.2) as well
as some internal component updates - new features or bug fixes in yandex-specific
components either on the master or nodes.

* `new_revision_summary` - Human readable description of the changes to be applied
when updating to the latest revision. Empty if new_revision_available is false.
* `version_deprecated` - Boolean flag. The current version is on the deprecation schedule,
component (master or node group) should be upgraded.

## Timeouts

This resource provides the following configuration options for 
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 15 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.

## Import

A Managed Kubernetes cluster can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_kubernetes_cluster.default cluster_id
```
