---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_image"
description: |-
  Creates a VM image for the Yandex Compute service from an existing tarball.
---

# yandex_compute_image (Resource)

Creates a virtual machine image resource for the Yandex Compute Cloud service from an existing tarball. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/image).

## Example usage

```terraform
//
// Create a new Compute Image.
//
resource "yandex_compute_image" "foo-image" {
  name       = "my-custom-image"
  source_url = "https://storage.yandexcloud.net/lucky-images/kube-it.img"
}

// You can use "data.yandex_compute_image.my_image.id" identifier 
// as reference to existing resource.
resource "yandex_compute_instance" "vm" {
  name = "vm-from-custom-image"

  # ...

  boot_disk {
    initialize_params {
      image_id = yandex_compute_image.foo-image.id
    }
  }
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

~> One of `source_family`, `source_image`, `source_snapshot`, `source_disk` or `source_url` must be specified.

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

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_compute_image.<resource Name> <resource Id>
terraform import yandex_compute_image.my_image fd8go**********trjsd
```
