//
// Get information about existing Compute Instance.
//
data "yandex_compute_instance" "my_instance" {
  instance_id = "some_instance_id"
}

output "instance_external_ip" {
  value = data.yandex_compute_instance.my_instance.network_interface.0.nat_ip_address
}
