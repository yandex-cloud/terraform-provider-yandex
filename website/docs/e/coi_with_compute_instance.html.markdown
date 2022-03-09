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

Read more [documentation](https://cloud.yandex.com/docs/container-registry/concepts/coi) about Container Optimized Image.

## Creating Container Optimized Image configuration with Terraform

This example shows how to create a simple project with a single instance based on Container Optimized Image from scratch.

First, create a Terraform config file named `main.tf`. Inside, you'll want to include the configuration of
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
* The `folder` field is the id of your folder to create Container Optimized Image.
* The `zone` field should be replaced with default [availability zone](https://cloud.yandex.com/docs/overview/concepts/geo-scope) to operate under.

Use already created Container Optimized Image from [image family](https://cloud.yandex.com/docs/compute/concepts/images#family) collection :

```hcl
data "yandex_compute_image" "container-optimized-image" {
  family    = "container-optimized-image"
}
```

Create compute instance:

```hcl
resource "yandex_compute_instance" "instance-based-on-coi" {

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.container-optimized-image.id
    }
  }
  network_interface {
    subnet_id = "your subnet id"
    nat       = true
  }
  resources {
    cores  = 2
    memory = 2
  }

  metadata = {
    docker-container-declaration = file("${path.module}/declaration.yaml")
    user-data                    = file("${path.module}/cloud_config.yaml")
  }
}
```

Configure compute instance:

* The `subnet_id` field is the id of your virtual private cloud [subnet](https://www.terraform.io/docs/providers/yandex/d/datasource_vpc_subnet.html).

Create a cloud specification file named  `cloud-config.yaml` and put it to the same folder:

```yaml
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

* Fill the `ssh_authorized_keys` value with your public ssh key.

Create Container Optimized Image specification file named `declaration.yaml` and put it to the same folder:

```yaml
#declaration.yaml
spec:
  containers:
  - image: cr.yandex/yc/demo/coi:v1
    securityContext:
      privileged: false
    stdin: false
    tty: false
```

Create `output.tf` file to get the IP address of the Container Optimized Image:

```hcl
output "external_ip" {
  value = yandex_compute_instance.instance-based-on-coi.network_interface.0.nat_ip_address
}
```

## Launching Container Optimized Image

Now everything is set to launch the COI instance in Terraform. Execute the following list of instructions:

* Run `terraform plan`, then `terraform apply`.

* After `terraform apply` you will have public IPv4 address in the outputs:

    ```
    Outputs:

    external_ip = <some_IPv4>
    ```
* Access newly created virtual machine:

    ```shell
    ssh yc-user@<some_IPv4>
    ```

* Make http request to your virtual machine:

    ```shell
    curl <some_IPv4>
    ```

    You will get in the response:

    ```html
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

## Creating Instance Group with Container Optimized Image

This example shows how to create an instance group of Container Optimized Images.

Use [Yandex.Cloud Provider](https://www.terraform.io/docs/providers/yandex/index.html) and [compute image](https://www.terraform.io/docs/providers/yandex/d/datasource_compute_image.html)
from the previous examples showing the creation of Container Optimized Image with compute instance.
Use cloud specification in `cloud-config.yaml` file and container specification in `declaration.yaml` file.

Create Instance Group:

```hcl
resource "yandex_compute_instance_group" "ig-with-coi" {
  name               = "ig with coi"
  folder_id          = "your folder"
  service_account_id = "your service account id"
  instance_template {
    platform_id = "standard-v1"
    resources {
      memory = 2
      cores  = 1
    }
    boot_disk {
      mode = "READ_WRITE"
      initialize_params {
        image_id = data.yandex_compute_image.container-optimized-image.id
      }
    }
    network_interface {
      network_id = "your network id"
      subnet_ids = ["all your subnet ids"]
    }

    metadata = {
      docker-container-declaration = file("${path.module}/declaration.yaml")
      user-data = file("${path.module}/cloud_config.yaml")
    }
    service_account_id = "The ID of the service account authorized for this instance"
  }

  scale_policy {
    fixed_scale {
      size = 3
    }
  }

  allocation_policy {
    zones = ["all your availability zones"]
  }

  deploy_policy {
    max_unavailable = 2
    max_creating    = 2
    max_expansion   = 2
    max_deleting    = 2
  }
}
```

Configure Instance Group:

* The `name` field is your instance group name.
* The `folder_id` field is the id of your folder to create Container Optimized Image.
* The `service_account_id` field is the id of your [service account](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts) authorized for this instance group.
* The `network_id` field is the id of your cloud [network](https://cloud.yandex.com/docs/vpc/concepts/network#network).
* The `subnet_id` field is an array of your [subnet ids](https://cloud.yandex.com/docs/vpc/concepts/network#subnet).
* The `zones` field is an array of your [availability zones](https://cloud.yandex.com/docs/overview/concepts/geo-scope).
* The `instance_template.service_account_id` field is the id of your [service account](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts) authorized for this instance.
