data "yandex_logging_group" "my_group" {
  group_id = "some_yandex_logging_group_id"
}

output "log_group_retention_period" {
  value = data.yandex_logging_group.my_group.retention_period
}
