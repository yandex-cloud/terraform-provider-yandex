---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_group"
sidebar_current: "docs-yandex-datasource-organizationmanager-group"
description: |-
  Get information about a Yandex.Cloud Group.
---

# yandex\_organizationmanager\_group

Get information about a Yandex.Cloud Organization Manager Group. For more information, see
[the official documentation](https://cloud.yandex.com/en-ru/docs/organization/manage-groups).

## Example Usage

```hcl
data "yandex_organizationmanager_group" group {
  group_id        = "some_group_id"
  organization_id = "some_organization_id"
}

output "my_group.name" {
  value = "${data.yandex_organizationmanager_group.group.name}"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) ID of a Group.

* `name` - (Optional) Name of a Group.

~> **NOTE:** One of `group_id` or `name` should be specified.

* `organization_id` - (Optional) Organization that the Group belongs to. If value is omitted, the default provider organization is used.

## Attributes Reference

The following attributes are exported:

* `description` - The description of the Group.
* `created_at` - The Group creation timestamp.
* `members` - A list of members of the Group. The structure is documented below.

The `members` block supports:
* `id` - The ID of the member.
* `type` - The type of the member.
