//
// Get information about an existing DataLens dashboard.
//
data "yandex_datalens_dashboard" "my-dashboard" {
  id              = "example-dashboard-id"
  organization_id = "example-organization-id"
}
