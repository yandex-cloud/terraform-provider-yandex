//
// Create a new Application Load Balancer (ALB)
//
resource "yandex_alb_load_balancer" "my_alb" {
  name = "my-load-balancer"

  network_id = yandex_vpc_network.test-network.id

  allocation_policy {
    location {
      zone_id   = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.test-subnet.id
    }
  }

  listener {
    name = "my-listener"
    endpoint {
      address {
        external_ipv4_address {
        }
      }
      ports = [8080]
    }
    http {
      handler {
        http_router_id = yandex_alb_http_router.test-router.id
      }
    }
  }

  log_options {
    discard_rule {
      http_code_intervals = ["2XX"]
      discard_percent     = 75
    }
  }
}
