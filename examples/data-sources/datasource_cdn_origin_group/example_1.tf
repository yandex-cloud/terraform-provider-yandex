data "yandex_cdn_origin_group" "my_group" {
  origin_group_id = "some_instance_id"
}

output "origin_group_name" {
  value = data.yandex_cdn_origin_group.my_group.name
}
