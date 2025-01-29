---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Organization Manager Group Mapping Items.
---

# {{.Name}} ({{.Type}})

{{ .Description }}

~> Group mapping items depends on group mapping. If you create group mapping via terraform use "depends_on" meta-argument to avoid errors (see example below).

## Example Usage

{{ tffile "examples/organizationmanager_group_mapping_item/r_organizationmanager_group_mapping_item_1.tf" }}

{{ .SchemaMarkdown }}

## Import

Resource can be imported using the following syntax:

{{ codefile "shell" "examples/organizationmanager_group_mapping_item/import.sh" }}

