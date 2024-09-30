---
subcategory: "Datasphere"
page_title: "Yandex: yandex_datasphere_community"
description: |-
  Allows management of a Yandex.Cloud Datasphere Community.
---


# yandex_datasphere_community




Allows management of Yandex Cloud Datasphere Communities

```terraform
resource "yandex_datasphere_project_iam_binding" "project-iam" {
  project_id = "your-datasphere-project-id"
  role       = "datasphere.community-projects.developer"
  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) Organization ID where community would be created
* `name` - (Required) Name of the Datasphere Community.
* `description` - (Optional) Datasphere Community description.
* `labels` - (Optional) A set of key/value label pairs to assign to the Datasphere Community.
* `billing_account_id` - (Optional) Billing account ID to associated with community

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Datasphere Community unique identifier
* `created_at` - Creation timestamp of the Yandex Datasphere Community
* `created_by` - Creator account ID of the Yandex Datasphere Community

## Timeouts

This resource provides the following configuration options for timeouts:

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

A Datasphere Community can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_datasphere_community.default community_id
```
