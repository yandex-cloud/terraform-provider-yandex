data "yandex_cloud_desktops_desktop" "data_desktop_by_name" {
	name = "desktop-group-name"
	folder_id = "<your folder id (optional)>"
}

data "yandex_cloud_desktops_desktop" "data_desktop_by_name" {
	desktop_id = "<your desktop id>"
}
