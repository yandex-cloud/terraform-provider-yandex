---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a VPC Security Group Rule within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{tffile "yandex-framework/docs-templates/vpc_security_group_rule/resource-example-1.tf"}}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/vpc_security_group_rule/resource-import.sh" }}
