//
// Create a new ALB HTTP Router
//
resource "yandex_alb_http_router" "tf-router" {
  name = "my-http-router"
  labels {
    tf-label    = "tf-label-value"
    empty-label = "s"
  }
}
