//
// Get information about Trino cluster by name
//
data "yandex_tirno_cluster" "trino_cluster_by_name" {
  name = "trino-created-with-terraform"
}

//
// Get information about Trino cluster by id
//
data "yandex_trino_cluster" "tirno_cluster_by_id" {
  id = "<trino-cluster-id>"
}
