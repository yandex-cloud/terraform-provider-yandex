//
// Get information about existing DNS Inbound Endpoint.
//
data "yandex_dns_inbound_endpoint" "foo" {
  dns_inbound_endpoint_id = yandex_dns_inbound_endpoint.endpoint1.id
}

output "address" {
  value = data.yandex_dns_inbound_endpoint.foo.address
}
