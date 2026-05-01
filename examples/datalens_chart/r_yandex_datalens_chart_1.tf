//
// Create a DataLens wizard chart. The chart payload mirrors the DataLens API:
// `annotation { description }` and `data { ... }` (visualization, placeholders,
// wizard|ql variant block, etc.).
//
resource "yandex_datalens_chart" "my-chart" {
  name            = "example-chart"
  organization_id = "example-organization-id"
  workbook_id     = "example-workbook-id"

  annotation = {
    description = "Sales by country"
  }

  data = {
    visualization = {
      id = "line"
      placeholders = [
        { id = "x", items = [] },
        { id = "y", items = [] },
      ]
    }
    wizard = {
      datasets_ids = [yandex_datalens_dataset.my-dataset.id]
    }
  }
}
