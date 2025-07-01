//
// Create a new Object Storage binding with extended partitioning.
//

resource "yandex_yq_object_storage_binding" "my_os_binding3" {
  name          = "tf-test-os-binding3"
  description   = "Binding has been created from Terraform"
  connection_id = yandex_yq_object_storage_connection.my_os_connection.id
  compression   = "gzip"
  format        = "json_each_row"

  partitioned_by = [
    "date",
    "severity",
  ]
  path_pattern = "/cold"
  projection = {
    "projection.date.format"     = "/year=%Y/month=%m/day=%d"
    "projection.date.interval"   = "1"
    "projection.date.max"        = "NOW"
    "projection.date.min"        = "2022-12-01"
    "projection.date.type"       = "date"
    "projection.date.unit"       = "DAYS"
    "projection.enabled"         = "true"
    "projection.severity.type"   = "enum"
    "projection.severity.values" = "error,info,fatal"
    "storage.location.template"  = "/$${date}/$${severity}"
  }

  column {
    name     = "timestamp"
    not_null = false
    type     = "String"
  }
  column {
    name     = "message"
    not_null = false
    type     = "String"
  }
  column {
    name     = "date"
    not_null = true
    type     = "Date"
  }
  column {
    name     = "severity"
    not_null = true
    type     = "String"
  }
}
