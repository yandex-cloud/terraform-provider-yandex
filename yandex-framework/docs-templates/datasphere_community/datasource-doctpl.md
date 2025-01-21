---
subcategory: "Datasphere"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Cloud Datasphere Community.
---


# {{.Name}}

{{ .Description }}


Get information about a Yandex Cloud Datasphere Community.

## Example usage

{{tffile "yandex-framework/docs-templates/datasphere_community/datasource-example-1.tf"}}

This data source is used to define Yandex Cloud Datasphere Community that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `community_id` - (Required) Yandex Cloud Datasphere Community id used to define community

## Attributes Reference

The following attributes are exported:

* `organization_id` - Organization ID where community would be created
* `name` - Name of the Datasphere Community.
* `description` - Datasphere Community description.
* `labels` - A set of key/value label pairs to assign to the Datasphere Community.
* `billing_account_id` - Billing account ID to associated with community
* `created_at` - Creation timestamp of the Yandex Datasphere Community
* `created_by` - Creator account ID of the Yandex Datasphere Community
