---
layout: "yandex"
page_title: "Yandex: yandex_compute_image"
sidebar_current: "docs-yandex-compute-image"
description: |-
  Creates a VM image for Yandex Compute Service from an existing tarball.
---

# yandex\_compute\_image

Creates a virtual machine image resource for Yandex Compute Cloud service from an existing
tarball. For more information see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/images).


## Example Usage

```hcl
resource "yandex_compute_image" "foo-image" {
  name       = "my-custom-image"
  source_url = "https://storage.yandexcloud.net/lucky-images/kube-it.img"
}

resource "yandex_compute_instance" "vm" {
  name = "vm-from-custom-image"
  ...

  boot_disk {
    initialize_params {
      image_id = "${yandex_compute_image.foo-image.id}"
    }
  }
}
```

## Argument Reference

The following arguments are supported: (Note that one of either source_image, source_snapshot,
  source_disk or source_url is required)

* `min_disk_size` - Minimum size in Gb of the disk that will be created from this image.

* `os_type` - Operating system type that is contained in the image. Possible values: "LINUX", "WINDOWS".

- - -

* `name` - (Optional) Name of the disk.

* `description` - (Optional) An optional description of the image. Provide this property when
  you create the resource.

* `family` - (Optional) The name of the image family to which this image belongs.

* `labels` - (Optional) A set of key/value label pairs to assign to the image.

* `source_family` - (Optional) The name of the family to find the ID of the latest image in "standard-images" folder, that will be used as the source of the image. Changing this forces a new resource to be created.

* `source_image` - (Optional) The ID of an image that will be used as the source of the
    image. Changing this forces a new resource to be created.

* `source_snapshot` - (Optional) The ID of a snapshot that will be used as the source of the
    image. Changing this forces a new resource to be created.

* `source_disk` - (Optional) The ID of a disk that will be used as the source of the
    image. Changing this forces a new resource to be created.

* `source_url` - (Optional) The URL that will be used as the source of the
    image. Changing this forces a new resource to be created.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

~> **NOTE:** One of `source_family`, `source_image`, `source_snapshot`, `source_disk` or `source_url` must be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `size` - The size of the image, specified in GB.
* `status` - The status of the image.

## Timeouts

`yandex_compute_image` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - Default 5 minutes
- `update` - Default 5 minutes
- `delete` - Default 5 minutes

## Import

VM image can be imported using the `id` of resource, e.g.

```
$ terraform import yandex_compute_image.web-image id
```
