//
// Get information about existing Compute Reserved Instance Pool.
//
data "yandex_compute_reserved_instance_pool" "pool" {
  reserved_instance_pool_id = "pool-id"
}