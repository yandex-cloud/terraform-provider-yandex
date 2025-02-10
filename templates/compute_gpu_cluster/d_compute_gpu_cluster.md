---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute GPU cluster.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute GPU cluster. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/gpu-cluster).

## Example usage

{{ tffile "examples/compute_gpu_cluster/d_compute_gpu_cluster_1.tf" }}

## Argument Reference

The following arguments are supported:

* `gpu_cluster_id` - (Optional) ID of the GPU cluster.

* `name` - (Optional) Name of the GPU cluster.

~> One of `gpu_cluster_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - Optional description of the GPU cluster.
* `folder_id` - ID of the folder that the GPU cluster belongs to.
* `zone` - ID of the zone where the GPU cluster resides.
* `interconnect_type` - type of interconnect used between nodes in GPU cluster.
* `status` - Current status of the GPU cluster.
* `labels` - GPU cluster labels as `key:value` pairs. For details about the concept, see [documentation](https://yandex.cloud/docs/overview/concepts/services#labels).
* `created_at` - Creation timestamp.
