---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Disk Placement Group resource.
---

# {{.Name}} ({{.Type}})

A Disk Placement Group resource. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/disk#nr-disks).

## Example usage

{{ tffile "examples/compute_disk_placement_group/r_compute_disk_placement_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

* `name` - (Optional) The name of the Disk Placement Group.

* `description` - (Optional) A description of the Disk Placement Group.

* `labels` - (Optional) A set of key/value label pairs to assign to the Disk Placement Group.

* `zone` - ID of the zone where the Disk Placement Group resides. Default is `ru-central1-b`

* `status` - Status of the Disk Placement Group.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_disk_placement_group/import.sh" }}
