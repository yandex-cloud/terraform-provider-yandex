---
subcategory: "Managed Service for Apache Airflow"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages an Apache Airflow cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/airflow_cluster/r_airflow_cluster_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/airflow_cluster/import.sh" }}
