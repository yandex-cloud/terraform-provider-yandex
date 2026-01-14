---
subcategory: "Managed Service for Greenplum"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Greenplum cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/mdb_greenplum_cluster_v2/r_mdb_greenplum_cluster_v2_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_greenplum_cluster_v2/import.sh" }}
