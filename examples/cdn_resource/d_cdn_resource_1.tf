//
// Get information about existing CDN Resource by resource_id
//
data "yandex_cdn_resource" "my_resource" {
  resource_id = "some_resource_id"
}

output "cdn_cname" {
  value = data.yandex_cdn_resource.my_resource.cname
}

//
// Get information about existing CDN Resource by cname
//
data "yandex_cdn_resource" "by_cname" {
  cname = "cdn.example.com"
}

output "cdn_origin_group_id" {
  value = data.yandex_cdn_resource.by_cname.origin_group_id
}
