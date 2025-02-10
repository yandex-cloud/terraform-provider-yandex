//
// Create a new ALB Backend Group.
//
resource "yandex_alb_backend_group" "my_alb_bg" {
  name = "my-backend-group"

  session_affinity {
    connection {
      source_ip = "127.0.0.1"
    }
  }

  http_backend {
    name             = "test-http-backend"
    weight           = 1
    port             = 8080
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    tls {
      sni = "backend-domain.internal"
    }
    load_balancing_config {
      panic_threshold = 50
    }
    healthcheck {
      timeout  = "1s"
      interval = "1s"
      http_healthcheck {
        path = "/"
      }
    }
    http2 = "true"
  }
}
