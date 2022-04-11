---
layout: "yandex"
page_title: "Yandex: yandex_compute_instance"
sidebar_current: "docs-yandex-compute-instance-x"
description: |-
  Manages a VM instance resource.
---

# yandex\_compute\_instance

A VM instance resource. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/vm).

## Example Usage

```hcl
resource "yandex_compute_instance" "default" {
  name        = "test"
  platform_id = "standard-v1"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = "image_id"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.foo.id}"
    nat = true
    security_group_ids = [ yandex_vpc_security_group.security_group1.id ]
  }

  metadata = {
    foo      = "bar"
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
    user-data = "${file("/home/user/cloudinit/meta.txt")}"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone       = "ru-central1-a"
  network_id = "${yandex_vpc_network.foo.id}"
}

resource "yandex_iam_service_account" "savpcsg" {                       
  name = "savpcsg"                                                      
}                                                                       
                                                                        
resource "yandex_resourcemanager_folder_iam_binding" "editor" {         
  folder_id = var.yc_folder_id                                    
  role      = "editor"                                                  
  members = [                                                           
    "serviceAccount:${yandex_iam_service_account.savpcsg.id}"           
  ]                                                                     
  depends_on = [                                                        
    yandex_iam_service_account.savpcsg,                                 
  ]                                                                     
}

resource "yandex_vpc_security_group" "security_group1" {                
  name        = "Security Group 1"                                      
  description = "Our Security Group"              
  network_id  = yandex_vpc_network.foo.id                           
  depends_on = [                                                        
    yandex_iam_service_account.savpcsg,                                 
    yandex_resourcemanager_folder_iam_binding.editor,                   
    yandex_vpc_network.foo,                                         
    yandex_vpc_subnet.foo,                            
  ]                                                                     
                                                                        
  ingress {                                                             
    protocol       = "TCP"                                              
    description    = "SSH"                                              
    port           = 22                                                 
    v4_cidr_blocks = ["0.0.0.0/0"]                                      
  }                                                                     
                                                                        
  egress {                                                              
                                                                        
    protocol       = "ANY"                                              
    description    = "Outbound unrestricted"                            
    v4_cidr_blocks = ["0.0.0.0/0"]                                      
    port           = -1                                                 
  }                                                                     
                                                                        
}
```
```
#Sample cloudinit/meta.txt
users:
  - name: user1
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    homedir: /opt/user1
    ssh-authorized-keys:
      - ssh-rsa .... <comment>

packages:
  - tmux
  - rsync

package_upgrade: true
package_reboot_if_required: true
timezone: 'Asia/Tokyo'

runcmd:
  - echo PermitRootLogin No >> /etc/ssh/sshd_config && sshd -t &&  systemctl restart sshd
  - echo -e "net.core.rmem_max=4194304\nnet.core.wmem_max=1048576" > /etc/sysctl.d/mynetwork.conf
  - sysctl -p /etc/sysctl.d/mynetwork.conf
  - chown -R user1:user1 /opt/user1
```

## Argument Reference

The following arguments are supported:

* `resources` - (Required) Compute resources that are allocated for the instance. The structure is documented below.

* `boot_disk` - (Required) The boot disk for the instance. The structure is documented below.

* `network_interface` - (Required) Networks to attach to the instance. This can
    be specified multiple times. The structure is documented below.

- - -

* `name` - (Optional) Resource name.

* `description` - (Optional) Description of the instance.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance.

* `zone` - (Optional) The availability zone where the virtual machine will be created. If it is not provided,
    the default provider folder is used.

* `hostname` - (Optional) Host name for the instance. This field is used to generate the instance `fqdn` value. 
    The host name must be unique within the network and region. If not specified, the host name will be equal 
    to `id` of the instance and `fqdn` will be `<id>.auto.internal`. 
    Otherwise FQDN will be `<hostname>.<region_id>.internal`.                        

* `metadata` - (Optional) Metadata key/value pairs to make available from
    within the instance.

* `platform_id` - (Optional) The type of virtual machine to create. The default is 'standard-v1'.

* `secondary_disk` - (Optional) A list of disks to attach to the instance. The structure is documented below.
    **Note**: The [`allow_stopping_for_update`](#allow_stopping_for_update) property must be set to true in order to update this structure.

* `scheduling_policy` - (Optional) Scheduling policy configuration. The structure is documented below.

* `placement_policy` - (Optional) The placement policy configuration. The structure is documented below.

* `service_account_id` - (Optional) ID of the service account authorized for this instance.

* `allow_stopping_for_update` - (Optional) If true, allows Terraform to stop the instance in order to update its properties.
    If you try to update a property that requires stopping the instance without setting this field, the update will fail.
    
* `network_acceleration_type` - (Optional) Type of network acceleration. The default is `standard`. Values: `standard`, `software_accelerated`

---

The `resources` block supports:

* `cores` - (Required) CPU cores for the instance.

* `memory` - (Required) Memory size in GB.

* `core_fraction` - (Optional) If provided, specifies baseline performance for a core as a percent.

The `boot_disk` block supports:

* `auto_delete` - (Optional) Defines whether the disk will be auto-deleted when the instance
    is deleted. The default value is `True`.

* `device_name` - (Optional) Name that can be used to access an attached disk.

* `mode` - (Optional) Type of access to the disk resource. By default, a disk is attached in `READ_WRITE` mode.

* `disk_id` - (Optional) The ID of the existing disk (such as those managed by
    `yandex_compute_disk`) to attach as a boot disk.

* `initialize_params` - (Optional) Parameters for a new disk that will be created
    alongside the new instance. Either `initialize_params` or `disk_id` must be set. The structure is documented below.

~> **NOTE:** Either `initialize_params` or `disk_id` must be specified.

The `initialize_params` block supports:

* `name` - (Optional) Name of the boot disk.

* `description` - (Optional) Description of the boot disk.

* `size` - (Optional) Size of the disk in GB.

* `type` - (Optional) Disk type.

* `image_id` - (Optional) A disk image to initialize this disk from.

* `snapshot_id` - (Optional) A snapshot to initialize this disk from.

~> **NOTE:** Either `image_id` or `snapshot_id` must be specified.

The `network_interface` block supports:

* `subnet_id` - (Required) ID of the subnet to attach this
    interface to. The subnet must exist in the same zone where this instance will be
    created.

* `ipv4` - (Optional) Allocate an IPv4 address for the interface. The default value is `true`.

* `ip_address` - (Optional) The private IP address to assign to the instance. If
    empty, the address will be automatically assigned from the specified subnet.

* `ipv6` - (Optional) If true, allocate an IPv6 address for the interface.
    The address will be automatically assigned from the specified subnet.

* `ipv6_address` - (Optional) The private IPv6 address to assign to the instance.

* `nat` - (Optional) Provide a public address, for instance, to access the internet over NAT.

* `nat_ip_address` - (Optional) Provide a public address, for instance, to access the internet over NAT. Address should be already reserved in web UI.

* `security_group_ids` - (Optional) Security group ids for network interface.

* `dns_record` - (Optional) List of configurations for creating ipv4 DNS records. The structure is documented below.

* `ipv6_dns_record` - (Optional) List of configurations for creating ipv6 DNS records. The structure is documented below.

* `nat_dns_record` - (Optional) List of configurations for creating ipv4 NAT DNS records. The structure is documented below.

The `dns_record` block supports:

* `fqdn` - (Required) DNS record FQDN (must have a dot at the end).

* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL. in seconds

* `ptr` - (Optional) When set to true, also create a PTR DNS record.

The `ipv6_dns_record` block supports:

* `fqdn` - (Required) DNS record FQDN (must have a dot at the end).

* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL. in seconds

* `ptr` - (Optional) When set to true, also create a PTR DNS record.

The `nat_dns_record` block supports:

* `fqdn` - (Required) DNS record FQDN (must have a dot at the end).

* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL. in seconds

* `ptr` - (Optional) When set to true, also create a PTR DNS record.

The `secondary_disk` block supports:

* `disk_id` - (Required) ID of the disk that is attached to the instance.

* `auto_delete` - (Optional) Whether the disk is auto-deleted when the instance
    is deleted. The default value is false.

* `device_name` - (Optional) Name that can be used to access an attached disk
    under `/dev/disk/by-id/`.

* `mode` - (Optional) Type of access to the disk resource. By default, a disk is attached in `READ_WRITE` mode.

The `scheduling_policy` block supports:

* `preemptible` - (Optional) Specifies if the instance is preemptible. Defaults to false.

The `placement_policy` block supports:

* `placement_group_id` - (Optional) Specifies the id of the Placement Group to assign to the instance.



## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `fqdn` - The fully qualified DNS name of this instance.

* `network_interface.0.ip_address` - The internal IP address of the instance.

* `network_interface.0.nat_ip_address` - The external IP address of the instance.

* `status` - The status of this instance.

* `created_at` - Creation timestamp of the instance.

## Import

Instances can be imported using the `ID` of an instance, e.g.

```
$ terraform import yandex_compute_instance.default instance_id
```
