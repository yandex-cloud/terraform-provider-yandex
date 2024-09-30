resource "yandex_function_scaling_policy" "my_scaling_policy" {
  function_id = "are1samplefunction11"
  policy {
    tag                  = "$latest"
    zone_instances_limit = 3
    zone_requests_limit  = 100
  }
  policy {
    tag                  = "my_tag"
    zone_instances_limit = 4
    zone_requests_limit  = 150
  }
}
