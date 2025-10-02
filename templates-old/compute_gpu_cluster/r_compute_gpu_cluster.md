---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  GPU Cluster connects multiple Compute GPU Instances in the same availability zone with high-speed low-latency network.
---

# {{.Name}} ({{.Type}})

GPU Cluster connects multiple Compute GPU Instances in the same availability zone with high-speed low-latency network.

Users can create a cluster from several VMs and use GPUDirectRDMA to directly send data between GPUs on different VMs.

For more information about GPU cluster in Yandex Cloud, see:

* [Documentation](https://yandex.cloud/docs/compute/concepts/gpu_cluster)

## Example usage

{{ tffile "examples/compute_gpu_cluster/r_compute_gpu_cluster_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the GPU cluster. Provide this property when you create a resource.

* `description` - (Optional) Description of the GPU cluster. Provide this property when you create a resource.

* `folder_id` - (Optional) The ID of the folder that the GPU cluster belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to assign to this GPU cluster. A list of key/value pairs. For details about the concept, see [documentation](https://yandex.cloud/docs/overview/concepts/services#labels).

* `zone` - (Optional) Availability zone where the GPU cluster will reside.

* `interconnect_type` - (Optional) Type of interconnect between nodes to use in GPU cluster. Type `INFINIBAND` is set by default, and it is the only one available at the moment.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the GPU cluster.
* `created_at` - Creation timestamp of the GPU cluster.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_gpu_cluster/import.sh" }}
