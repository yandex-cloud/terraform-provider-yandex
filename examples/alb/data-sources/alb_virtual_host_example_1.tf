data "yandex_alb_virtual_host" "my-vh-data" {
  name           = yandex_alb_virtual_host.my-vh.name
  http_router_id = yandex_alb_virtual_host.my-router.id
}
