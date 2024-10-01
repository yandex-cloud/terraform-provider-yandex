resource "yandex_compute_gpu_cluster" "cluster1" {
  name              = "gpu-cluster-name"
  interconnect_type = "infiniband"
  zone              = "ru-central1-a"

  labels = {
    environment = "test"
  }
}

resource "yandex_compute_gpu_cluster_iam_binding" "editor" {
  gpu_cluster_id = data.yandex_compute_gpu_cluster.cluster1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
