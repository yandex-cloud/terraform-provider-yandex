data "yandex_kubernetes_cluster" "my_cluster" {
  cluster_id = "some_k8s_cluster_id"
}

output "cluster_external_v4_endpoint" {
  value = data.yandex_kubernetes_cluster.my_cluster.master.0.external_v4_endpoint
}
