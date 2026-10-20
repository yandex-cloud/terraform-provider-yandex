// Create a private connection on an existing trunk with an allocated VLAN.
// ipv4_static_routes manages the full route set; use [] to remove all routes.
// Omitted routes retain their current values during updates.
// After import, configure peer_bgp_md5_key again if BGP authentication is used.
//
resource "yandex_cic_private_connection" "example" {
  name                = "my-private-connection"
  trunk_connection_id = data.yandex_cic_trunk_connection.example.id
  vlan_id             = 101

  ipv4_peering = {
    peering_subnet = "192.168.0.0/30"
    cloud_ip       = "192.168.0.1"
    peer_ip        = "192.168.0.2"
    peer_bgp_asn   = 65001
  }

  ipv4_static_routes = [
    { prefix = "10.10.0.0/24" },
  ]
}

data "yandex_cic_trunk_connection" "example" {
  trunk_connection_id = "trunk-connection-id"
}
