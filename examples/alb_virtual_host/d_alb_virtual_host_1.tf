//
// Get information about existing ALB Virtual Host
//
data "yandex_alb_virtual_host" "my-vhost" {
  name           = yandex_alb_virtual_host.my-vh.name
  http_router_id = yandex_alb_virtual_host.my-router.id
}
