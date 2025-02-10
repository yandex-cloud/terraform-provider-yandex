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
