---
subcategory: "Managed Service for MongoDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MongoDB Database within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_mongodb_database/r_mdb_mongodb_database_1.tf" }}

{{ .SchemaMarkdown | trimspace }}


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/mdb_mongodb_database/import.sh" }}
