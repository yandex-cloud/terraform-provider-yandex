//
// Get information about existing GPU Cluster.
//
data "yandex_compute_gpu_cluster" "my_gpu_cluster" {
  gpu_cluster_id = "some_gpu_cluster_id"
}

// You can use "data.yandex_compute_gpu_cluster.my_gpu_cluster.id" identifier 
// as reference to the existing resource.
resource "yandex_compute_instance" "default" {
  # ...

  gpu_cluster_id = data.yandex_compute_gpu_cluster.my_gpu_cluster.id

}
