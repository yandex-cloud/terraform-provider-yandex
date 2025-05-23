---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Cloud Group.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Cloud Organization Manager Group. For more information, see [the official documentation](https://yandex.cloud/docs/organization/manage-groups).

## Example usage

{{ tffile "examples/organizationmanager_group/d_organizationmanager_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) ID of a Group.

* `name` - (Optional) Name of a Group.

~> One of `group_id` or `name` should be specified.

* `organization_id` - (Optional) Organization that the Group belongs to. If value is omitted, the default provider organization is used.

## Attributes Reference

The following attributes are exported:

* `description` - The description of the Group.
* `created_at` - The Group creation timestamp.
* `members` - A list of members of the Group. The structure is documented below.

The `members` block supports:
* `id` - The ID of the member.
* `type` - The type of the member.
