//
// Create a new Network Load Balancer (NLB).
//
resource "yandex_lb_network_load_balancer" "my_nlb" {
  name = "my-network-load-balancer"

  listener {
    name = "my-listener"
    port = 8080
    external_address_spec {
      ip_version = "ipv4"
    }
  }

  attached_target_group {
    target_group_id = yandex_lb_target_group.my-target-group.id

    healthcheck {
      name = "http"
      http_options {
        port = 8080
        path = "/ping"
      }
    }
  }
}
