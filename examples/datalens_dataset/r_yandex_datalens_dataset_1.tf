//
// Create a DataLens dataset on top of a YDB connection.
//
resource "yandex_datalens_dataset" "my-dataset" {
  name            = "example-dataset"
  organization_id = "example-organization-id"
  workbook_id     = "example-workbook-id"

  dataset = {
    description = "Example dataset on top of a YDB connection"

    sources = [{
      id            = "src-1"
      title         = "events"
      source_type   = "YDB_TABLE"
      connection_id = yandex_datalens_connection.my-conn.id
      parameters = {
        table_name = "events"
      }
    }]

    source_avatars = [{
      id        = "ava-1"
      source_id = "src-1"
      title     = "events"
      is_root   = true
    }]

    result_schema = [
      {
        guid      = "f-country"
        title     = "Country"
        data_type = "string"
        type      = "DIMENSION"
        avatar_id = "ava-1"
        source    = "country"
      },
      {
        guid        = "f-revenue-sum"
        title       = "Revenue (sum)"
        data_type   = "float"
        type        = "MEASURE"
        avatar_id   = "ava-1"
        source      = "revenue"
        aggregation = "sum"
      },
    ]
  }
}
