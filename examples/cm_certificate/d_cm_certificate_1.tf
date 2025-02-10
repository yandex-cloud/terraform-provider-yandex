//
// Get information about existing CM Certificate
//
data "yandex_cm_certificate" "example_by_id" {
  certificate_id = "certificate-id"
}

data "yandex_cm_certificate" "example_by_name" {
  folder_id = "folder-id"
  name      = "example"
}
