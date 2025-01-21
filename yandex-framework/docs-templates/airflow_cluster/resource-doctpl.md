---
subcategory: "Managed Service for Apache Airflow"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages an Apache Airflow cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "yandex-framework/docs-templates/airflow_cluster/resource-example-1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/airflow_cluster/resource-import.sh" }}
