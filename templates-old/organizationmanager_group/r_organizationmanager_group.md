---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single Group within an existing Yandex Cloud Organization.
---

# {{.Name}} ({{.Type}})

Allows management of a single Group within an existing Yandex Cloud Organization. For more information, see [the official documentation](https://yandex.cloud/docs/organization/manage-groups).

## Example usage

{{ tffile "examples/organizationmanager_group/r_organizationmanager_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Group.
* `description` - (Optional) The description of the Group.
* `organization_id` - (Required, Forces new resource) The organization to attach this Group to.

## Attributes Reference

* `created_at` - (Computed) The SAML Federation creation timestamp.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/organizationmanager_group/import.sh" }}
