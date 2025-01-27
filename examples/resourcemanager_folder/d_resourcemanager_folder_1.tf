# Get folder by ID
data "yandex_resourcemanager_folder" "my_folder_1" {
  folder_id = "folder_id_number_1"
}

# Get folder by name in specific cloud
data "yandex_resourcemanager_folder" "my_folder_2" {
  name     = "folder_name"
  cloud_id = "some_cloud_id"
}

output "my_folder_1_name" {
  value = data.yandex_resourcemanager_folder.my_folder_1.name
}

output "my_folder_2_cloud_id" {
  value = data.yandex_resourcemanager_folder.my_folder_2.cloud_id
}

