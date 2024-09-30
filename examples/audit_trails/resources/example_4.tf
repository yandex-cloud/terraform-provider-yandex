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
