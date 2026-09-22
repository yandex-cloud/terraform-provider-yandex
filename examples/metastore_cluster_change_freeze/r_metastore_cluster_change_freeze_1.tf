resource "yandex_metastore_cluster_change_freeze" "example" {
  resource_id = yandex_metastore_cluster.example.id
  start_at    = "2026-12-30T00:00:00Z"
  end_at      = "2027-01-06T00:00:00Z"
  reason      = "New Year holidays"
}
