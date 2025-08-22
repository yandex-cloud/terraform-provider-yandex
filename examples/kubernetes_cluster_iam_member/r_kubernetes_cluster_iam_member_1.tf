//
// Create a new Managed Kubernetes cluster and a new IAM Member for it.
//
data "yandex_kubernetes_cluster" "my_cluster" {
  name = "My Managed Kubernetes cluster"
}

resource "yandex_kubernetes_cluster_iam_member" "viewer" {
  cluster_id = yandex_kubernetes_cluster.my_cluster.id
  role = "viewer"
  member = "userAccount:my_user_account_id"
}
