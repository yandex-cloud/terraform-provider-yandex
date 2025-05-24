//
// Create a new YDS binding.
//

resource "yandex_yq_yds_binding" "my_yds_binding" {
    name = "tf-test-os-binding"
    description = "Binding has been created from Terraform"
    connection_id = yandex_yq_yds_connection.my_yds_connection.id
    format = "csv_with_names"
    stream = "my_stream"
    format_setting = {
      "data.datetime.format_name" = "POSIX"
    }
    column {
        name = "ts"
        type = "Timestamp"
    }
    column {
        name = "message"
        type = "utf8"
    }
}
