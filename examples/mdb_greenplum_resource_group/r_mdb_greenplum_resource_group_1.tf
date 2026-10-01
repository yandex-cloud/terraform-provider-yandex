// Create a resource group in an existing Apache Cloudberry cluster.
resource "yandex_mdb_greenplum_resource_group" "analytics" {
  cluster_id      = "<cloudberry-cluster-id>"
  name            = "analytics"
  concurrency     = 10
  cpu_max_percent = 80
  cpu_weight      = 100
  memory_quota    = 1024
  min_cost        = 100
}
