//
// Create a new column-oriented YDB Table.
//
resource "yandex_ydb_table" "test_table" {
  path              = "test_dir/test_table_3_col"
  connection_string = yandex_ydb_database_serverless.database1.ydb_full_endpoint

  column {
    name     = "a"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "b"
    type     = "Uint32"
    not_null = true
  }
  column {
    name     = "c"
    type     = "Int32"
    not_null = false
  }
  column {
    name = "d"
    type = "Timestamp"
  }

  primary_key = ["a", "b"]

  store = "column"

  partitioning_settings {
    partition_by = ["b", "a"]
  }
}
