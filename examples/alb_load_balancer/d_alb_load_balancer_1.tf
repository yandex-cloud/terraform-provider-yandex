//
// Get information about existing Application Load Balancer (ALB).
//
data "yandex_alb_load_balancer" "tf-alb-data" {
  load_balancer_id = "my-alb-id"
}
