//
// Get information about an existing DataLens workbook.
//
data "yandex_datalens_workbook" "my-workbook" {
  id              = "example-workbook-id"
  organization_id = "example-organization-id"
}
