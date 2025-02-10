//
// Create a new IoT Core Registry.
//
resource "yandex_iot_core_registry" "my_registry" {
  name        = "some_name"
  description = "any description"
  labels = {
    my-label = "my-label-value"
  }
  log_options {
    log_group_id = "log-group-id"
    min_level    = "ERROR"
  }
  passwords = [
    "my-password1",
    "my-password2"
  ]
  certificates = [
    "public part of certificate1",
    "public part of certificate2"
  ]
}
