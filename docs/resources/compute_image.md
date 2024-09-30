---
subcategory: "Compute"
page_title: "Yandex: yandex_compute_image"
description: |-
  Creates a VM image for the Yandex Compute service from an existing tarball.
---


# yandex_compute_image




Creates a virtual machine image resource for the Yandex Compute Cloud service from an existing tarball. For more information, see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/image).

```terraform
resource "yandex_compute_snapshot_schedule" "schedule1" {
  schedule_policy {
    expression = "0 0 * * *"
  }

  retention_period = "12h"

  snapshot_spec {
    description = "retention-snapshot"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}

resource "yandex_compute_snapshot_schedule_iam_binding" "editor" {
  snapshot_schedule_id = data.yandex_compute_snapshot_schedule.schedule1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the disk.

* `description` - (Optional) An optional description of the image. Provide this property when you create a resource.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the image.

* `family` - (Optional) The name of the image family to which this image belongs.

* `min_disk_size` - (Optional) Minimum size in GB of the disk that will be created from this image.

* `os_type` - (Optional) Operating system type that is contained in the image. Possible values: "LINUX", "WINDOWS".

* `pooled` - (Optional) Optimize the image to create a disk.

* `source_family` - (Optional) The name of the family to use as the source of the new image. The ID of the latest image is taken from the "standard-images" folder. Changing the family forces a new resource to be created.

* `source_image` - (Optional) The ID of an existing image to use as the source of the image. Changing this ID forces a new resource to be created.

* `source_snapshot` - (Optional) The ID of a snapshot to use as the source of the image. Changing this ID forces a new resource to be created.

* `source_disk` - (Optional) The ID of a disk to use as the source of the image. Changing this ID forces a new resource to be created.

* `source_url` - (Optional) The URL to use as the source of the image. Changing this URL forces a new resource to be created.

* `product_ids` - (Optional) License IDs that indicate which licenses are attached to this image.

* `hardware_generation` - (Optional) Hardware generation and its features,
  which will be applied to the instance when this image is used as a boot
  disk source. Provide this property if you wish to override this value, which
  otherwise is inherited from the source. The structure is documented below.

~> **NOTE:** One of `source_family`, `source_image`, `source_snapshot`, `source_disk` or `source_url` must be specified.

The `hardware_generation` consists of one of the following blocks:

* `legacy_features` - Defines the first known hardware generation and its features, which are:
  * `pci_topology` - A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.
* `generation2_features` - A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `size` - The size of the image, specified in GB.
* `status` - The status of the image.
* `created_at` - Creation timestamp of the image.

## Timeouts

`yandex_compute_image` provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 5 minutes
- `update` - Default 5 minutes
- `delete` - Default 5 minutes

## Import

A VM image can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_compute_image.web-image image_id
```
