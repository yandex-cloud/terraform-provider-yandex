---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: {{.Name}}"
description: |-
  Virtual hosts combine routes belonging to the same set of domains.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/alb_virtual_host/r_alb_virtual_host_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

The `resource ID` for the ALB virtual host is defined as its `http router id` separated by `/` from the `virtual host's name`.

{{ codefile "bash" "examples/alb_virtual_host/import.sh" }}
