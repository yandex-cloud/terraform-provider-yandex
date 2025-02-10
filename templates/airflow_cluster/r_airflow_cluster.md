---
subcategory: "Managed Service for Apache Airflow"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages an Apache Airflow cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ codefile "terraform" "examples/airflow_cluster/r_airflow_cluster_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/airflow_cluster/import.sh" }}
