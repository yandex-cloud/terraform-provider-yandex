data "yandex_cdn_resource" "my_resource" {
  resource_id = "some resource id"
}

output "resource_cname" {
  value = data.yandex_cdn_resource.my_resource.cname
}
