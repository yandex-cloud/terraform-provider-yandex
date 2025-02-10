//
// Get information about existing ALB HTTP Router
//
data "yandex_alb_http_router" "tf-router" {
  http_router_id = "my-http-router-id"
}
