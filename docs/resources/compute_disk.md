---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_disk"
description: |-
  Persistent disks are durable storage devices that function similarly to the physical disks in a desktop or a server.
---

# yandex_compute_disk (Resource)

Persistent disks are used for data storage and function similarly to physical hard and solid state drives.

A disk can be attached or detached from the virtual machine and can be located locally. A disk can be moved between virtual machines within the same availability zone. Each disk can be attached to only one virtual machine at a time.

For more information about disks in Yandex Cloud, see:
* [Documentation](https://yandex.cloud/docs/compute/concepts/disk)
* How-to Guides:
  * [Attach and detach a disk](https://yandex.cloud/docs/compute/concepts/disk#attach-detach)
  * [Backup operation](https://yandex.cloud/docs/compute/concepts/disk#backup)

~> Only one of `image_id` or `snapshot_id` can be specified.

## Example usage

```terraform
//
// Create a new Compute Disk.
//
resource "yandex_compute_disk" "my_disk" {
  name     = "disk-name"
  type     = "network-ssd"
  zone     = "ru-central1-a"
  image_id = "ubuntu-16.04-v20180727"

  labels = {
    environment = "test"
  }
}
```

```terraform
//
// Create a new Compute Disk and put it to the specific Placement Group.
//
resource "yandex_compute_disk" "my_vm" {
  name = "non-replicated-disk-name"
  size = 93 // Non-replicated SSD disk size must be divisible by 93G
  type = "network-ssd-nonreplicated"
  zone = "ru-central1-b"

  disk_placement_policy {
    disk_placement_group_id = yandex_compute_disk_placement_group.my_pg.id
  }
}

resource "yandex_compute_disk_placement_group" "my_pg" {
  zone = "ru-central1-b"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `allow_recreate` (Boolean)
- `block_size` (Number) Block size of the disk, specified in bytes.
- `description` (String) The resource description.
- `disk_placement_policy` (Block List, Max: 1) Disk placement policy configuration. (see [below for nested schema](#nestedblock--disk_placement_policy))
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `hardware_generation` (Block List, Max: 1) Hardware generation and its features, which will be applied to the instance when this disk is used as a boot disk. Provide this property if you wish to override this value, which otherwise is inherited from the source. (see [below for nested schema](#nestedblock--hardware_generation))
- `image_id` (String) The source image to use for disk creation.
- `kms_key_id` (String) ID of KMS symmetric key used to encrypt disk.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `name` (String) The resource name.
- `size` (Number) Size of the persistent disk, specified in GB. You can specify this field when creating a persistent disk using the `image_id` or `snapshot_id` parameter, or specify it alone to create an empty persistent disk. If you specify this field along with `image_id` or `snapshot_id`, the size value must not be less than the size of the source image or the size of the snapshot.
- `snapshot_id` (String) The source snapshot to use for disk creation.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `type` (String) Type of disk to create. Provide this when creating a disk.
- `zone` (String) The [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `id` (String) The ID of this resource.
- `product_ids` (List of String)
- `status` (String) The status of the disk.

<a id="nestedblock--disk_placement_policy"></a>
### Nested Schema for `disk_placement_policy`

Required:

- `disk_placement_group_id` (String) Specifies Disk Placement Group id.


<a id="nestedblock--hardware_generation"></a>
### Nested Schema for `hardware_generation`

Optional:

- `generation2_features` (Block List, Max: 1) A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot. (see [below for nested schema](#nestedblock--hardware_generation--generation2_features))
- `legacy_features` (Block List, Max: 1) Defines the first known hardware generation and its features. (see [below for nested schema](#nestedblock--hardware_generation--legacy_features))

<a id="nestedblock--hardware_generation--generation2_features"></a>
### Nested Schema for `hardware_generation.generation2_features`


<a id="nestedblock--hardware_generation--legacy_features"></a>
### Nested Schema for `hardware_generation.legacy_features`

Optional:

- `pci_topology` (String) A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_compute_disk.<resource Name> <resource Id>
terraform import yandex_compute_disk.my_disk fhmrm**********90r5f
```
