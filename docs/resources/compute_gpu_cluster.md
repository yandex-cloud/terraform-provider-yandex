---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_gpu_cluster"
description: |-
  GPU Cluster connects multiple Compute GPU Instances in the same availability zone with high-speed low-latency network.
---

# yandex_compute_gpu_cluster (Resource)

GPU Cluster connects multiple Compute GPU Instances in the same availability zone with high-speed low-latency network.

Users can create a cluster from several VMs and use GPUDirect RDMA to directly send data between GPUs on different VMs.

For more information about GPU cluster in Yandex Cloud, see:
* [Documentation](https://yandex.cloud/docs/compute/concepts/gpu_cluster)

## Example usage

```terraform
//
// Create a new GPU Cluster.
//
resource "yandex_compute_gpu_cluster" "default" {
  name              = "gpu-cluster-name"
  interconnect_type = "infiniband"
  zone              = "ru-central1-a"

  labels = {
    environment = "test"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `interconnect_type` (String) Type of interconnect between nodes to use in GPU cluster. Type `INFINIBAND` is set by default, and it is the only one available at the moment.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `name` (String) The resource name.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `zone` (String) The [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `id` (String) The ID of this resource.
- `status` (String) The status of the GPU cluster.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_compute_gpu_cluster.<resource Name> <resource Id>
terraform import yandex_compute_gpu_cluster.my_gpu_cluster fv4h4**********u4dpa
```
