data "yandex_organizationmanager_group" "group" {
  group_id        = "some_group_id"
  organization_id = "some_organization_id"
}

output "my_group.name" {
  value = data.yandex_organizationmanager_group.group.name
}
