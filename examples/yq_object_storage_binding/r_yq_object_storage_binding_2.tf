//
// Create a new Object Storage binding with Hive partitioning.
//

resource "yandex_yq_object_storage_binding" "my_os_binding2" {
  name          = "tf-test-os-binding2"
  description   = "Binding has been created from Terraform"
  connection_id = yandex_yq_object_storage_connection.my_os_connection.id
  format        = "csv_with_names"
  path_pattern  = "my_logs/"
  format_setting = {
    "file_pattern" = "*.csv"
  }
  column {
    name     = "year"
    type     = "Int32"
    not_null = true
  }
  column {
    name     = "month"
    type     = "Int32"
    not_null = true
  }
  column {
    name     = "day"
    type     = "Int32"
    not_null = true
  }

  partitioned_by = ["year", "month", "day"]
  column {
    name = "ts"
    type = "Timestamp"
  }
  column {
    name = "message"
    type = "Utf8"
  }
}
