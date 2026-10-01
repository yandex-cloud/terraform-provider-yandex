//
// Export management events to Object Storage with a one-minute aggregation period.
//
resource "yandex_audit_trails_trail" "storage_trail" {
  name               = "storage-trail"
  folder_id          = "home-folder"
  service_account_id = "trail-service-account"

  storage_destination {
    bucket_name        = "audit-events-bucket"
    object_prefix      = "events/"
    aggregation_period = "1m"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
  }
}
