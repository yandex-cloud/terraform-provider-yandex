//
// Get information about existing YTsaurus cluster.
//
data "yandex_ytsaurus_cluster" "my_cluster" {
  cluster_id = "some_cluster_id"
}