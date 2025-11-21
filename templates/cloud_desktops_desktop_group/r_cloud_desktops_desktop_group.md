---
subcategory: "Cloud Desktops"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Cloud Desktops Group.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ codefile "terraform" "examples/cloud_desktops_desktop_group/r_cloud_desktops_desktop_group_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `name`, `folderID`, `desktopImageID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/cloud_desktops_desktop_group/import.sh" }}
