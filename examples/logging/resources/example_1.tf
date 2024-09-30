resource "yandex_logging_group" "group1" {
  name      = "test-logging-group"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id
}
