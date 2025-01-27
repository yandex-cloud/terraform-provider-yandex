---
subcategory: "Client Config"
page_title: "Yandex: {{.Name}}"
description: |-
  Get attributes used by provider to configure client connection.
---

# {{.Name}} ({{.Type}})

Get attributes used by provider to configure client connection.

## Example usage

{{ tffile "examples/client_config/d_client_config_1.tf" }}

## Attributes Reference

The following attributes are exported:

* `cloud_id` - The ID of the cloud that the provider is connecting to.
* `folder_id` - The ID of the folder in which we operate.
* `zone` - The default availability zone.
* `iam_token` - A short-lived token that can be used for authentication in a Kubernetes cluster.
