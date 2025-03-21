---
subcategory: "Managed Service for YDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Manage a YDB Topic.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/ydb_topic/r_ydb_topic_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/ydb_topic/import.sh" }}
