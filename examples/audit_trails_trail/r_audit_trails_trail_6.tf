//
// Export management events to the same Monium project as Audit Trails metrics.
// The destination project cannot be configured.
//
resource "yandex_audit_trails_trail" "monium_trail" {
  name               = "monium-trail"
  folder_id          = "home-folder"
  service_account_id = "trail-service-account"

  monium_destination {}

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
  }
}
