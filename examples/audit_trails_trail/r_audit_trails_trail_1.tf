//
// Create a new basic Audit Trails Trail
//
resource "yandex_audit_trails_trail" "basic-trail" {
  name        = "basic-trail"
  folder_id   = "home-folder"
  description = "Some trail description"

  labels = {
    key = "value"
  }

  service_account_id = "trail-service-account"

  logging_destination {
    log_group_id = "some-log-group"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    data_events_filter {
      service = "storage"
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    data_events_filter {
      service = "dns"
      resource_scope {
        resource_id   = "vpc-net-id-1"
        resource_type = "vpc.network"
      }
      resource_scope {
        resource_id   = "vpc-net-id-2"
        resource_type = "vpc.network"
      }
    }
  }
}
