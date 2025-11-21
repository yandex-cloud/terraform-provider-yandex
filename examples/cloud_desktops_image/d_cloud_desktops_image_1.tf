data "yandex_cloud_desktops_image" "desktop_image_by_id" {
	id = "fdvtmm38i0rp795kkkpa"
}

data "yandex_cloud_desktops_image" "desktop_image_by_folder_and_name" {
	folder_id = "<your folder id (optional)>"
	name 	  = "Ubuntu 20.04 LTS (2024-12-03)"
}
