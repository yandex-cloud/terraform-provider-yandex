//
// Get information about existing IP Permission of specific Cloud Registry.
//
data "yandex_cloudregistry_registry_ip_permission" "my_ip_permission_by_id" {
  registry_id = yandex_cloudregistry_registry.my_registry.id
}
