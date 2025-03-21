---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manage a Web Application Firewall in Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/sws_waf_profile/r_sws_waf_profile_1.tf" }}

{{ tffile "examples/sws_waf_profile/r_sws_waf_profile_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/sws_waf_profile/import.sh" }}
