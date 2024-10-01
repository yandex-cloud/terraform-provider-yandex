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
