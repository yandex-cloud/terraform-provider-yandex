// 
// Get CM Certificate payload. Can be used for Certificate Validation.
//
data "yandex_cm_certificate_content" "example_by_id" {
  certificate_id = "certificate-id"
}

data "yandex_cm_certificate_content" "example_by_name" {
  folder_id = "folder-id"
  name      = "example"
}
