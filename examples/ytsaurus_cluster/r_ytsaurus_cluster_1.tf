//
// Create a new YTsaurus cluster
//
resource "yandex_ytsaurus_cluster" "my_cluster" {
  name = "my-cluster"
  description = "my_cluster description"

  zone_id			 = "ru-central1-a"
  subnet_id			 = "my_subnet_id"
  security_group_ids = ["my_security_group_id"]

  spec = {
	storage = {
	  hdd = {
	  	size_gb = 100
		count 	= 3
	  }

	  ssd = {
	  	size_gb = 100
		type 	= "network-ssd"
		count 	= 3
	  }
	}

	compute = [{
	  preset = "c8-m32"
	  disks = [{
	  	type 	= "network-ssd"
		size_gb = 50
	  }]
	  scale_policy = {
	  	fixed = {
		  size = 3
		}
	  }
	}]

	tablet = {
      preset = "c8-m16"
	  count = 3
	}

	proxy = {
	  http = {
	  	count = 1
	  }
      
	  rpc = {
	  	count = 1
	  }
	}
  }
}