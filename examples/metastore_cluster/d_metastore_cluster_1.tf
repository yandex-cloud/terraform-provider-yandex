//
// Get information about Metastore Cluster by name
//
data "yandex_metastore_cluster" "metastore_cluster_by_name" {
  name = "metastore-created-with-terraform"
}

//
// Get information about Metastore Cluster by id
//
data "yandex_metastore_cluster" "metastore_cluster_by_id" {
  id = "<metastore-cluster-id>"
}
