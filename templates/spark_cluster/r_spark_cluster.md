---
subcategory: "Managed Service for Apache Spark"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages an Apache Spark cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ codefile "terraform" "examples/spark_cluster/r_spark_cluster_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/spark_cluster/import.sh" }}
