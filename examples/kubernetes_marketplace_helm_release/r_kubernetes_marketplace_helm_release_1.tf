//
// Create a new Kubernetes Marketplace Helm Release. 
//
resource "yandex_kubernetes_marketplace_helm_release" "gatekeeper_helm_release" {
  cluster_id = yandex_kubernetes_cluster.cluster_resource_name.id

  product_version = "f2ecif2vt62k2637tgus" // Gatekeeper 3.12.0

  name      = "gatekeeper"
  namespace = kubernetes_namespace.namespace_resource_name.name

  user_values = {
    auditInterval             = "90"
    constraintViolationsLimit = "30"
  }
}
