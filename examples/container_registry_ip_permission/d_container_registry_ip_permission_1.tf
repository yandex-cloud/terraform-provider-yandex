//
// Get information about existing IP Permission of specific Container Registry.
//
data "yandex_container_registry_ip_permission" "my_ip_permission_by_id" {
  registry_id = yandex_container_registry.my_registry.id
}
