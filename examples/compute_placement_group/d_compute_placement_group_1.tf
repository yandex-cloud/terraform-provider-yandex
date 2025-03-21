//
// Get information about existing Compute Placement Group
//
data "yandex_compute_placement_group" "my_group" {
  group_id = "some_group_id"
}

output "placement_group_name" {
  value = data.yandex_compute_placement_group.my_group.name
}
