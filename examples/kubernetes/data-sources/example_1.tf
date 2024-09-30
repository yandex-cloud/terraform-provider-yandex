data "yandex_kubernetes_node_group" "my_node_group" {
  node_group_id = "some_k8s_node_group_id"
}

output "my_node_group.status" {
  value = data.yandex_kubernetes_node_group.my_node_group.status
}
