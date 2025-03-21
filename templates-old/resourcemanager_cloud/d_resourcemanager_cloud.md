---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Retrieve Yandex RM Cloud details.
---

# {{.Name}} ({{.Type}})

Use this data source to get cloud details. For more information, see [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#cloud).

## Example usage

{{ tffile "examples/resourcemanager_cloud/d_resourcemanager_cloud_1.tf" }}

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) ID of the cloud.
* `name` - (Optional) Name of the cloud.

~> Either `cloud_id` or `name` must be specified.

## Attributes Reference

The following attributes are returned:

* `name` - Name of the cloud.
* `description` - Description of the cloud.
* `created_at` - Cloud creation timestamp.
