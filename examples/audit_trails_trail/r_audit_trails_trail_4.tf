//
// Migration from deprecated filter field
//

// Before replacing "filter.path_filter block to the "filtering_policy.management_events_filter" block.
filter {
  path_filter {
    any_filter {
      resource_id   = "home-folder"
      resource_type = "resource-manager.folder"
    }
  }
}

// After replacing "filter.path_filter block to the "filtering_policy.management_events_filter" block.
filtering_policy {
  management_events_filter {
    resource_scope {
      resource_id   = "home-folder"
      resource_type = "resource-manager.folder"
    }
  }
}
