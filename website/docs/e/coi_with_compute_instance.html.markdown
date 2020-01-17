---
layout: "yandex"
page_title: "Yandex: coi_with_compute_instance"
sidebar_current: "docs-container-optimized-image"
description: |-
  Creating Container Optimized Image with Terraform.
---

# Container Optimized Image

Container Optimized Image is an [image](https://cloud.yandex.com/docs/compute/concepts/image) that is optimized for running Docker containers. 
The image includes Ubuntu LTS, Docker, and a daemon for launching Docker containers.

It's integrated with the Yandex.Cloud platform, which allows you to:

* Run a Docker container immediately after the VM is created from the management console or YC CLI.
* Update running Docker containers with minimum downtime.
* Access Docker container open network ports without any additional configuration.

Read more [documentation](https://cloud.yandex.com/docs/container-registry/concepts/coi) about container optimized image.

## Creating Container Optimized Image configuration with Terraform

This example shows how to create a simple project with a single container optimized image from scratch.
 
First create a Terraform config file named ```main.tf```. Inside, you'll want to include the configuration of 
[Yandex.Cloud Provider](https://www.terraform.io/docs/providers/yandex/index.html), 
[compute instance](https://www.terraform.io/docs/providers/yandex/r/compute_instance.html)
and [compute image](https://www.terraform.io/docs/providers/yandex/d/datasource_compute_image.html).

Use Yandex provider:

```hcl
provider "yandex" {
  token     = "your YC_TOKEN"
  folder_id = "your folder id"
  zone      = "your default zone"
}
```

Configure Yandex provider:

* The `token` field should be replaced with your personal Yandex.Cloud authentication token.
* The `folder` field is the id of your folder to create container optimized image.
* The `zone` field should be replaced with default availability zone operate under.

Use already created Container Optimized Image from [image family](https://cloud.yandex.com/docs/compute/concepts/images#family) collection :

```hcl
data "yandex_compute_image" "container-optimized-image" {
  family    = "container-optimized-image"
  folder_id = "standard-images"
}
```

Create compute instance:

```hcl
resource "yandex_compute_instance" "this" {

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.container-optimized-image.id}"
    }
  }
  network_interface {
    subnet_id = "your-subnet-id"
    nat       = true
  }
  resources {
    cores  = 2
    memory = 2
  }

  metadata = {
    docker-container-declaration = file("${path.module}/declaration.yaml")
    ec2-user-data                = file("${path.module}/cloud_config.yaml")
    user-data                    = file("${path.module}/cloud_config.yaml")
  }
}
```

Configure compute instance:

* The subnet_id field is the id of your virtual private cloud [subnet](https://www.terraform.io/docs/providers/yandex/d/datasource_vpc_subnet.html).

Create cloud specification file named  ```cloud-config.yaml``` and put it to the same folder:

```hcl
#cloud-config.yaml
ssh_pwauth: no
users:
  - name: yc-user
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - "your public ssh key"
```

Configure cloud specification:

* Fill the ssh_authorized_keys value with your public ssh key.

Create container optimized image specification file named ```declaration.yaml``` and put to the same folder:

```hcl
#declaration.yaml
spec:
  containers:
  - image: cr.yandex/yc/demo/coi:v1
    securityContext:
      privileged: false
    stdin: false
    tty: false
```

Create ```output.tf``` file to get IP address of the container optimized image:

```hcl
output "external_ip" {
  value = "${yandex_compute_instance.this.network_interface.0.nat_ip_address}"
}
```

## Launching Container Optimized Image

Now everything is set to launch the COI in Terraform. Make following list of instructions:

* Run ```terraform plan```, then ```terraform apply```.

* After ```terraform apply``` you will have public IPv4 address in the outputs:
  
  ```hcl
  Outputs:

  external_ip = <some_IPv4>
  ```
* Access newly created virtual machine:
  
  ```hcl
  ssh yc-user@<some_IPv4>
  ```

* Make http request to your virtual machine:
  
  ```hcl
  curl <some_IPv4>
  ```
  
  You will get in the response:
  
  ```hcl
  <!DOCTYPE html>
  <html lang="en">
  <head>
     <meta http-equiv="refresh" content="3">
     <title>Yandex.Scale</title>
  </head>
  <body>
  <h1>Hello v1</h1>
  </body>
  </html>
  ```