//
// Get information about a routing instance by ID.
//
data "yandex_cloudrouter_routing_instance" "example" {
  routing_instance_id = yandex_cloudrouter_routing_instance.example.id
}

output "routing_instance_status" {
  value = data.yandex_cloudrouter_routing_instance.example.status
}
