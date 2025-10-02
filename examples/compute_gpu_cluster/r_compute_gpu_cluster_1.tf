//
// Create a new GPU Cluster.
//
resource "yandex_compute_gpu_cluster" "default" {
  name              = "gpu-cluster-name"
  interconnect_type = "INFINIBAND"
  zone              = "ru-central1-a"

  labels = {
    environment = "test"
  }
}
