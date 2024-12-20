resource "yandex_organizationmanager_group_mapping_item" "group_mapping_item" {
  federation_id = "my-federation_id"
  internal_group_id = "my_internal_group_id"
  external_group_id = "my_external_group_id"

  depends_on = [yandex_organizationmanager_group_mapping.group_mapping]
}
