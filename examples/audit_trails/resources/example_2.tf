resource "yandex_audit_trails_trail" "basic_trail" {
  name        = "a-trail"
  folder_id   = "home-folder"
  description = "Some trail description"

  labels = {
    key = "value"
  }

  service_account_id = "trail-service-account"

  data_stream_destination {
    database_id = "some-database"
    stream_name = "some-stream"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "some-organization"
        resource_type = "organization-manager.organization"
      }
    }
    data_events_filter {
      service = "storage"
      resource_scope {
        resource_id   = "some-organization"
        resource_type = "organization-manager.organization"
      }
    }
    data_events_filter {
      service = "dns"
      resource_scope {
        resource_id   = "some-organization"
        resource_type = "organization-manager.organization"
      }
    }
  }
}
