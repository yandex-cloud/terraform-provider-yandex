---
subcategory: "Network Load Balancer (NLB)"
page_title: "Yandex: {{.Name}}"
description: |-
  A network load balancer is used to evenly distribute the load across cloud resources.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/lb_network_load_balancer/r_lb_network_load_balancer_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/lb_network_load_balancer/import.sh" }}
