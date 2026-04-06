//
// Create a new Kubernetes Marketplace Helm Release.
//
resource "yandex_kubernetes_marketplace_helm_release" "gwin_helm_release" {
  cluster_id = yandex_kubernetes_cluster.cluster_resource_name.id

  product_version = "f2e04077v04sobds7gkt" // Gwin v1.1.0

  name      = "gwin"
  namespace = kubernetes_namespace.namespace_resource_name.metadata[0].name

  user_values = {
    "controller.folderId" = yandex_resourcemanager_folder.folder_resource_name.id
    "controller.ycServiceAccount.workloadIdentityFederation.serviceAccountID" = yandex_iam_service_account.service_account_resource_name.id
    "controller.defaultBalancerSubnets" = yamlencode([
      yandex_vpc_subnet.subnet_resource_name_1.id,
      yandex_vpc_subnet.subnet_resource_name_2.id
    ])
  }
}
