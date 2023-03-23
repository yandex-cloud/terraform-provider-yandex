---
layout: "yandex"
page_title: "Yandex: yandex_compute_gpu_cluster"
sidebar_current: "docs-yandex-compute-gpu-cluster"
description: |-
GPU Cluster connects multiple Compute GPU Instances in the same availability zone with high-speed low-latency network.
---

# yandex\_compute\_gpu\_cluster

GPU Cluster connects multiple Compute GPU Instances in the same availability zone with high-speed low-latency network.

Users can create a cluster from several VMs and use GPUDirectRDMA to directly send data between GPUs on different VMs.

For more information about GPU cluster in Yandex.Cloud, see:

* [Documentation](https://cloud.yandex.com/docs/compute/concepts/gpu_cluster)

## Example Usage

```hcl
resource "yandex_compute_gpu_cluster" "default" {
  name               = "gpu-cluster-name"
  interconnect_type  = "infiniband"
  zone               = "ru-central1-a"

  labels = {
    environment = "test"
  }
}
```

## Argument Reference

The following arguments are supported:


* `name` - (Optional) Name of the GPU cluster. Provide this property when you create a resource.

* `description` - (Optional) Description of the GPU cluster. Provide this property when you create a resource.

* `folder_id` - (Optional) The ID of the folder that the GPU cluster belongs to. If it is not provided, the default 
   provider folder is used.

* `labels` - (Optional) Labels to assign to this GPU cluster. A list of key/value pairs. For details about the concept, 
  see [documentation](https://cloud.yandex.com/docs/overview/concepts/services#labels).

* `zone` - (Optional) Availability zone where the GPU cluster will reside.

* `interconnect_type` - (Optional) Type of interconnect between nodes to use in GPU cluster. Type `infiniband` is set by default, 
  and it is the only one available at the moment.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the GPU cluster.
* `created_at` - Creation timestamp of the GPU cluster.

## Timeouts

This resource provides the following configuration options for
[timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

A GPU cluster can be imported using any of these accepted formats:

```
$ terraform import yandex_compute_gpu_cluster.default gpu_cluster_id
```
