//
// Get information about existing CDN Rule by rule_id
//
data "yandex_cdn_rule" "my_rule" {
  resource_id = "bc851ft45fne********"
  rule_id     = 123456
}

output "rule_pattern" {
  value = data.yandex_cdn_rule.my_rule.rule_pattern
}

output "rule_weight" {
  value = data.yandex_cdn_rule.my_rule.weight
}

//
// Get information about existing CDN Rule by name
//
data "yandex_cdn_rule" "by_name" {
  resource_id = yandex_cdn_resource.my_resource.id
  name        = "redirect-old-urls"
}

output "rule_options" {
  value = data.yandex_cdn_rule.by_name.options
}
