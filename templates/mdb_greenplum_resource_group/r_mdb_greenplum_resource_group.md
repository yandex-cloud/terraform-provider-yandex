---
subcategory: "Managed Service for Greenplum"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Greenplum resource group within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/mdb_greenplum_resource_group/r_mdb_greenplum_resource_group_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_greenplum_resource_group/import.sh" }}
