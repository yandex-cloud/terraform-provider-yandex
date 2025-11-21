// network to function
resource "yandex_vpc_network" "network" {}

resource "yandex_vpc_subnet" "subnet" {
	zone 			= "ru-central1-a"
	network_id 		= yandex_vpc_network.network.id
	v4_cidr_blocks 	= ["10.1.0.0/24"]
}

data "yandex_cloud_desktops_image" "desktop_image_by_folder_and_name" {
	name 	  = "Ubuntu 20.04 LTS (2024-12-03)"
}

// desktop group
resource "yandex_cloud_desktops_desktop_group" "desktop_group" {
	name 		= "desktop-group-name"
	folder_id 	= "<your folder id (optional)>"
	image_id 	= data.yandex_cloud_desktops_image.image.id
	description = "Sample description"
	
	desktop_template = {
		resources = {
			cores 			= 4
			memory 			= 8
			core_fraction 	= 100
		}
		boot_disk = {
			initialize_params = {
				size = 24
				type = "SSD"
			}
		}
		data_disk = {
			initialize_params = {
				size = 16
				type = "HDD"
			}
		}
		network_interface = {
			network_id = yandex_vpc_network.network.id
			subnet_ids = ["${yandex_vpc_subnet.subnet.id}"]
		}
		group_config = {
			initialize_params = {
				min_ready_desktops 	= 1
				max_desktops_amount = 5
				desktop_type 		= "PERSISTENT"
				members				= [
					{
						id 		= "<your id>"
						type 	= "userAccount"
					}
				]
			}
		}
	}
		
	labels = {
    	label1 = "label1-value"
    	label2 = "label2-value"
  	}
}

// desktop
resource "yandex_cloud_desktops_desktop" "desktop" {
	name 				= "desktop-name"
	desktop_group_id 	= yandex_cloud_desktops_desktop_group.desktop_group.desktop_group_id
	
	network_interface = {
		subnet_id = yandex_vpc_subnet.subnet.id
	}

	members = [
		{
			subject_id 		= "bfblmuiaug62t0cki3dq"
			subject_type 	= "userAccount"
		},
	]

	labels = {
		label1 = "label1-value"
		label2 = "label2-value"
	}
}
