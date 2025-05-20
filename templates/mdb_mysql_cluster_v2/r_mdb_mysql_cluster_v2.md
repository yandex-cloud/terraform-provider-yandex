---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MySQL cluster within the Yandex Cloud. For more information, see the official documentation https://cloud.yandex.com/docs/managed-mysql/. How to connect to the DB https://yandex.cloud/docs/managed-mysql/quickstart#connect. To connect, use port 6432. The port number is not configurable.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_mysql_cluster_v2/r_mdb_mysql_cluster_v2_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/mdb_postgresql_cluster_v2/import.sh" }}
