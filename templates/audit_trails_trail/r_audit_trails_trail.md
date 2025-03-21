---
subcategory: "Audit Trails"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a trail resource.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/audit_trails_trail/r_audit_trails_trail_1.tf" }}

{{ tffile "examples/audit_trails_trail/r_audit_trails_trail_2.tf" }}

{{ tffile "examples/audit_trails_trail/r_audit_trails_trail_3.tf" }}

{{ tffile "examples/audit_trails_trail/r_audit_trails_trail_4.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/audit_trails_trail/import.sh" }}
