//
// Get information about Apache Airflow cluster by name
//
data "yandex_airflow_cluster" "airflow_cluster_by_name" {
  name = "airflow-created-with-terraform"
}

//
// Get information about Apache Airflow cluster by id
//
data "yandex_airflow_cluster" "airflow_cluster_by_id" {
  id = "<airflow-cluster-id>"
}
