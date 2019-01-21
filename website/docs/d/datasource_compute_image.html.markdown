---
layout: "yandex"
page_title: "Yandex: yandex_compute_image"
sidebar_current: "docs-yandex-datasource-compute-image"
description: |-
  Get information about a Yandex Compute image.
---

# yandex\_compute\_image

Get information about a Yandex Compute image. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/images).

## Example Usage

```hcl
data "yandex_compute_image" "my_image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "default" {
  ...

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.my_image.id}"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `image_id` - (Optional) The ID of a specific image.

* `family` - (Optional) The family name of an image. Used to search the latest image in a family.

~> **NOTE:** Either `image_id` or `family` must be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If a value is not
  provided, the default provider folder is used.

~> **NOTE:** If you specify `family` without `folder_id` then lookup takes place in the 'standard-images' folder.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `name` - The name of the image.
* `description` - An optional description of this image.
* `family` - The OS family name of the image.
* `min_disk_size` - Minimum size of the disk which is created from this image.
* `size` - The size of the image, specified in Gb.
* `status` - The status of the image.
* `product_ids` - License IDs that indicate which licenses are attached to this image.
* `os_type` - Operating system type that the image contains.
* `labels` - A map of labels applied to this image.
* `created_at` - Image creation timestamp.
