---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute image.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute image. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/image).

## Example usage

{{ tffile "examples/compute_image/d_compute_image_1.tf" }}

## Argument Reference

The following arguments are supported:

* `image_id` - (Optional) The ID of a specific image.

* `family` - (Optional) The family name of an image. Used to search the latest image in a family.

* `name` - (Optional) The name of the image.

~> Either `image_id`, `family` or `name` must be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

~> If you specify `family` without `folder_id` then lookup takes place in the 'standard-images' folder.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - An optional description of this image.
* `family` - The OS family name of the image.
* `min_disk_size` - Minimum size of the disk which is created from this image.
* `size` - The size of the image, specified in Gb.
* `status` - The status of the image.
* `product_ids` - License IDs that indicate which licenses are attached to this image.
* `os_type` - Operating system type that the image contains.
* `labels` - A map of labels applied to this image.
* `created_at` - Image creation timestamp.
* `pooled` - Optimize the image to create a disk.
* `hardware_generation` - Image hardware generation and its features. The structure is documented below.
* `kms_key_id` - ID of KMS symmetric key used to encrypt image.

---

The `hardware_generation` consists of one of the following blocks:

* `legacy_features` - Defines the first known hardware generation and its features, which are:
  * `pci_topology` - A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.
* `generation2_features` - A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.
