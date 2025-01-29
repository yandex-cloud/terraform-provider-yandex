---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a VPC Security Group Rule within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{tffile "examples/vpc_security_group_rule/r_vpc_security_group_rule_1.tf"}}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/vpc_security_group_rule/import.sh" }}
