//
// Create a new Managed Kubernetes cluster and a new IAM Binding for it.
//
data "yandex_kubernetes_cluster" "my_cluster" {
  name = "My Managed Kubernetes cluster"
}

resource "yandex_kubernetes_cluster_iam_binding" "viewer" {
  cluster_id = yandex_kubernetes_cluster.my_cluster.id

  role = "viewer"

  members = [
    "userAccount:my_user_account_id",
    "serviceAccount:my_service_account_id",
  ]
}
