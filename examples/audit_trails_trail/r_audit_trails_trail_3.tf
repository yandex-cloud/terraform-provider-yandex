//
// Migration from deprecated filter field
//

// Before replacing "filter.event_filters.path_filter" to the "resource_scope" block.
event_filters {
  path_filter {
    some_filter {
      resource_id   = "home-folder"
      resource_type = "resource-manager.folder"
      any_filters {
        resource_id   = "vpc-net-id-1"
        resource_type = "vpc.network"
      }
      any_filters {
        resource_id   = "vpc-net-id-2"
        resource_type = "vpc.network"
      }
    }
  }
}

// After replacing "filter.event_filters.path_filter" to the "resource_scope" block.
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
