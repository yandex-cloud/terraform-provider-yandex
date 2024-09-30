resource "yandex_iot_core_broker" "my_broker" {
  name        = "some_name"
  description = "any description"
  labels = {
    my-label = "my-label-value"
  }
  log_options {
    log_group_id = "log-group-id"
    min_level    = "ERROR"
  }
  certificates = [
    "public part of certificate1",
    "public part of certificate2"
  ]
}
