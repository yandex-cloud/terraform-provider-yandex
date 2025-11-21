data "yandex_cloud_desktops_desktop_group" "data_desktop_group_by_name_and_folder" {
	name 		= "desktop-group-name"
	folder_id 	= "<your folder id (optional)>"
}

data "yandex_cloud_desktops_desktop_group" "data_desktop_group_by_id" {
	desktop_group_id 	= "<your desktop group id>"
}
