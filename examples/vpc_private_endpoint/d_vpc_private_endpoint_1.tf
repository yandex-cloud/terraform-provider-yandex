//
// Get information about existing VPC Private Endpoint.
//
data "yandex_vpc_private_endpoint" "pe" {
  private_endpoint_id = "my-private-endpoint-id"
}
