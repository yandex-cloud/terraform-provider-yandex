//
// Get information about an existing DataLens dataset.
//
data "yandex_datalens_dataset" "my-dataset" {
  id              = "example-dataset-id"
  organization_id = "example-organization-id"
}
