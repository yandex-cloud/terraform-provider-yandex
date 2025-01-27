data "yandex_compute_gpu_cluster" "my_gpu_cluster" {
  gpu_cluster_id = "some_gpu_cluster_id"
}

resource "yandex_compute_instance" "default" {
  ...

  gpu_cluster_id = data.yandex_compute_gpu_cluster.my_gpu_cluster.id

}
