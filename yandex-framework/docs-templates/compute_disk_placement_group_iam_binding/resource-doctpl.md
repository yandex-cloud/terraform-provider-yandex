---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for the Disk Placement Group.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "yandex-framework/docs-templates/compute_disk_placement_group_iam_binding/resource-example-1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/compute_disk_iam_binding/resource-import.sh" }}
```
