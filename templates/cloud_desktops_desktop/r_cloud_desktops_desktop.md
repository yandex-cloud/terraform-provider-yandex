---
subcategory: "Cloud Desktops"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Cloud Desktops Desktop.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ codefile "terraform" "examples/cloud_desktops_desktop/r_cloud_desktops_desktop_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `desktopID` and `subnetID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/cloud_desktops_desktop/import.sh" }}
