---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Organization Manager Group Mapping Items.
---

# {{.Name}} ({{.Type}})

{{ .Description }}

NOTE: Group mapping items depends on [group mapping](organizationmanager_group_mapping.html). If you create group mapping via terraform use "depends_on" meta-argument to avoid errors (see example below).

## Example Usage

{{ tffile "examples/organizationmanager_group_mapping_item/r_organizationmanager_group_mapping_item_1.tf" }}

{{ .SchemaMarkdown }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/organizationmanager_group_mapping_item/import.sh" }}

