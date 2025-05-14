//
// Get information about Apache Spark cluster by name
//
data "yandex_spark_cluster" "spark_cluster_by_name" {
  name = "spark-created-with-terraform"
}

//
// Get information about Apache Spark cluster by id
//
data "yandex_spark_cluster" "spark_cluster_by_id" {
  id = "<spark-cluster-id>"
}
