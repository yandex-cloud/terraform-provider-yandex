---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of members of Yandex Cloud Organization Manager Group.
---

# {{.Name}} ({{.Type}})

Allows members management of a single Yandex Cloud Organization Manager Group. For more information, see [the official documentation](https://yandex.cloud/docs/organization/manage-groups#add-member).

~> Multiple `yandex_organizationmanager_group_iam_binding` resources with the same group id will produce inconsistent behavior!

## Example usage

{{ tffile "examples/organizationmanager_group_membership/r_organizationmanager_group_membership_1.tf" }}

## Argument Reference

The following arguments are supported:

* `group_id` - (Required, Forces new resource) The Group to add/remove members to/from.
* `members` - A set of members of the Group. Each member is represented by an id.

## Import

~> Import for this resource is not implemented yet.

