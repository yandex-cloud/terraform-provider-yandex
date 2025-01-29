---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for the compute Filesystem.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/compute_filesystem_iam_binding/r_compute_filesystem_iam_binding_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/compute_disk_iam_binding/import.sh" }}
