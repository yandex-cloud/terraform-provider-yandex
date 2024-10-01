resource "yandex_lb_target_group" "foo" {
  name      = "my-target-group"
  region_id = "ru-central1"

  target {
    subnet_id = yandex_vpc_subnet.my-subnet.id
    address   = yandex_compute_instance.my-instance-1.network_interface.0.ip_address
  }

  target {
    subnet_id = yandex_vpc_subnet.my-subnet.id
    address   = yandex_compute_instance.my-instance-2.network_interface.0.ip_address
  }
}
