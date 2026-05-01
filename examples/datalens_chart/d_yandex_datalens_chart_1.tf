//
// Get information about an existing DataLens chart.
//
data "yandex_datalens_chart" "my-chart" {
  id              = "example-chart-id"
  type            = "wizard"
  organization_id = "example-organization-id"
}
