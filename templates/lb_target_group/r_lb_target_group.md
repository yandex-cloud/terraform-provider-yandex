---
subcategory: "Network Load Balancer (NLB)"
page_title: "Yandex: {{.Name}}"
description: |-
  A load balancer distributes the load across cloud resources that are combined into a target group.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/lb_target_group/r_lb_target_group_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/lb_target_group/import.sh" }}
