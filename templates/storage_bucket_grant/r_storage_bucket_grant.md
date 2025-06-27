---
subcategory: "Object Storage (S3)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of grants on an existing Yandex Cloud Storage Bucket.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/storage_bucket_grant/r_storage_bucket_grant_0.tf" }}

{{ tffile "examples/storage_bucket_grant/r_storage_bucket_grant_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/storage_bucket_grant/import.sh" }}
