---
subcategory: "Client Config"
page_title: "Yandex: yandex_client_config"
description: |-
  Get attributes used by provider to configure client connection.
---

# yandex_client_config (Data Source)

Get attributes used by provider to configure client connection.

## Example usage

```terraform
//
// Example of using Yandex Cloud client configuration
//
data "yandex_client_config" "client" {}

data "yandex_kubernetes_cluster" "kubernetes" {
  name = "kubernetes"
}

provider "kubernetes" {
  load_config_file = false

  host                   = data.yandex_kubernetes_cluster.kubernetes.master.0.external_v4_endpoint
  cluster_ca_certificate = data.yandex_kubernetes_cluster.kubernetes.master.0.cluster_ca_certificate
  token                  = data.yandex_client_config.client.iam_token
}
```

## Attributes Reference

The following attributes are exported:

* `cloud_id` - The ID of the cloud that the provider is connecting to.
* `folder_id` - The ID of the folder in which we operate.
* `zone` - The default availability zone.
* `iam_token` - A short-lived token that can be used for authentication in a Kubernetes cluster.
