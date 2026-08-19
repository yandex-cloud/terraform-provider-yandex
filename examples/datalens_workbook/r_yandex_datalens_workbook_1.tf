//
// Create a DataLens workbook at the organization root.
//
resource "yandex_datalens_workbook" "my-workbook" {
  title           = "Sales analytics"
  description     = "Workbook for sales BI artifacts"
  organization_id = "example-organization-id"
}
