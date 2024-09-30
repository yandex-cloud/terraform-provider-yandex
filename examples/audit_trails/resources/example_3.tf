resource "yandex_audit_trails_trail" "basic_trail" {
  name        = "a-trail"
  folder_id   = "home-folder"
  description = "Some trail description"

  labels = {
    key = "value"
  }

  service_account_id = "trail-service-account"

  storage_destination {
    bucket_name   = "some-bucket"
    object_prefix = "some-prefix"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    data_events_filter {
      service = "mdb.postgresql"
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
  }
}
