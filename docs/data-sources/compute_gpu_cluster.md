---
subcategory: "Compute"
page_title: "Yandex: yandex_compute_gpu_cluster"
description: |-
  Get information about a Yandex Compute GPU cluster.
---


# yandex_compute_gpu_cluster




Get information about a Yandex Compute GPU cluster. For more information, see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/gpu-cluster).

```terraform
data "yandex_compute_snapshot_schedule" "my_snapshot_schedule" {
  snapshot_schedule_id = "some_snapshot_schedule_id"
}
```

## Argument Reference

The following arguments are supported:

* `gpu_cluster_id` - (Optional) ID of the GPU cluster.

* `name` - (Optional) Name of the GPU cluster.

~> **NOTE:** One of `gpu_cluster_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - Optional description of the GPU cluster.
* `folder_id` - ID of the folder that the GPU cluster belongs to.
* `zone` - ID of the zone where the GPU cluster resides.
* `interconnect_type` - type of interconnect used between nodes in GPU cluster.
* `status` - Current status of the GPU cluster.
* `labels` - GPU cluster labels as `key:value` pairs. For details about the concept, see [documentation](https://cloud.yandex.com/docs/overview/concepts/services#labels).
* `created_at` - Creation timestamp.
