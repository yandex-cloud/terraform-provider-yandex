---
subcategory: "Cloud Content Delivery Network (CDN)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Yandex CDN Resource Rule.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/cdn_rule/r_cdn_rule_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

CDN rules can be imported using the composite ID in the format `resource_id/rule_id`, e.g.:

{{ codefile "shell" "examples/cdn_rule/import.sh" }}
