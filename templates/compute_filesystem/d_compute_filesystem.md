---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute filesystem.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute filesystem. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/filesystem).

## Example usage

{{ tffile "examples/compute_filesystem/d_compute_filesystem_1.tf" }}

## Argument Reference

The following arguments are supported:

* `filsystem_id` - (Optional) ID of the filesystem.

* `name` - (Optional) Name of the filesystem.

~> One of `filesystem_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - Optional description of the filesystem.
* `folder_id` - ID of the folder that the filesystem belongs to.
* `zone` - ID of the zone where the filesystem resides.
* `size` - Size of the filesystem, specified in Gb.
* `block_size` - The block size of the filesystem in bytes.
* `type` - ID of the filesystem type.
* `status` - Current status of the filesystem.
* `labels` - Filesystem labels as `key:value` pairs. For details about the concept, see [documentation](https://yandex.cloud/docs/overview/concepts/services#labels).
* `created_at` - Creation timestamp.
