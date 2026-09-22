resource "yandex_mdb_clickhouse_cluster_change_freeze" "example" {
  resource_id = yandex_mdb_clickhouse_cluster.example.id
  start_at    = "2026-11-26T00:00:00Z"
  end_at      = "2026-12-01T00:00:00Z"
  reason      = "Black Friday"
}
