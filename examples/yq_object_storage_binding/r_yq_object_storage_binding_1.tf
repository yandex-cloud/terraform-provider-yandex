//
// Create a new Object Storage binding.
//

resource "yandex_yq_object_storage_binding" "my_os_binding1" {
  name          = "tf-test-os-binding1"
  description   = "Binding has been created from Terraform"
  connection_id = yandex_yq_object_storage_connection.my_os_connection.id
  compression   = "gzip"
  format        = "json_each_row"

  path_pattern = "my_logs/"
  column {
    name = "ts"
    type = "Timestamp"
  }
  column {
    name = "message"
    type = "Utf8"
  }
}
