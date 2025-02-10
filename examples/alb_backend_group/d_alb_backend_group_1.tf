//
// Get information about existing ALB Backend Group
//
data "yandex_alb_backend_group" "my_alb_bg" {
  backend_group_id = yandex_alb_backend_group.my_backend_group.id
}
