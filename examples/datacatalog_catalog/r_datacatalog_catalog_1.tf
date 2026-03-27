resource "yandex_datacatalog_catalog" "tf_test_catalog" {
  name = "tf-test-catalog"
  folder_id = "folder_id"
  description = "test-catalog description"
  labels = {
    "label1" = "value0"
    "label2" = "value2"
    "label3" = "value3"
  }
}