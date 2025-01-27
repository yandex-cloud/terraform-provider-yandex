---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of the Cloud resource.
---

# {{.Name}} ({{.Type}})

Allows creation and management of Cloud resources for an existing Yandex Cloud Organization. See [the official documentation](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy) for additional info. Note: deletion of clouds may take up to 30 minutes as it requires a lot of communication between cloud services.

## Example usage

{{ tffile "examples/resourcemanager_cloud/r_resourcemanager_cloud_1.tf" }}

## Argument Reference

The following arguments are supported:

* `organization_id` - (Optional) Yandex Cloud Organization that the cloud belongs to. If value is omitted, the default provider Organization ID is used.

* `name` - (Optional) The name of the Cloud.

* `description` - (Optional) A description of the Cloud.

* `labels` - (Optional) A set of key/value label pairs to assign to the Cloud.
