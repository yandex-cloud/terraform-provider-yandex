//
// Get information about existing VPC Private Endpoint.
//
data "yandex_vpc_private_endpoint" "pe" {
  private_endpoint_id = "my-private-endpoint-id"
}

//
// Use dns_records to get the DNS record FQDN assigned to the private endpoint.
//
output "dns_record_fqdn" {
  value = data.yandex_vpc_private_endpoint.pe.dns_records[0].name
}
