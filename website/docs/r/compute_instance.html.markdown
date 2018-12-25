---
layout: "yandex"
page_title: "Yandex: yandex_compute_instance"
sidebar_current: "docs-yandex-compute-instance-x"
description: |-
  Manages a VM instance resource.
---

# yandex\_compute\_instance

VM instance resource. For more information see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/vm)

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
  }

  metadata {
    foo      = "bar"
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }
}

resource "yandex_vpc_network" "foo" {
}

resource "yandex_vpc_subnet" "foo" {
    zone       = "ru-central1-a"
    network_id = "${yandex_vpc_network.foo.id}"
}

```

## Argument Reference

The following arguments are supported:

* `resources` - (Required) Compute resources that are allocated for instance.
    The structure is documented below.

* `boot_disk` - (Required) The boot disk for the instance.
    The structure is documented below.

* `network_interface` - (Required) Networks to attach to the instance. This can
    be specified multiple times. The structure is documented below.

- - -

* `name` - (Optional) Resource name.

* `description` - (Optional) Description of the instance.

* `zone` - (Optional) The availability zone where the machine will be created. If it is not provided,
    the default provider folder is used.

* `platform_id` - (Optional) Type of the virtual machine to create. Default is 'standard-v1'.

* `secondary_disk` - (Optional) A list of disks to attach to the instance. The structure is documented below.
    **Note**: [`allow_stopping_for_update`](#allow_stopping_for_update) must be set to true in order to update this structure.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance.

* `metadata` - (Optional) Metadata key/value pairs to make available from
    within the instance.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `allow_stopping_for_update` - (Optional)  If true, allows Terraform to stop the instance to update its properties.
    If you try to update a property that requires stopping the instance without setting this field, the update will fail.

---

The `resources` block supports:

* `cores` - (Required) CPU cores for instance.

* `memory` - (Required) Memory size in Gb.

* `core_fraction` - (Optional) If provided, specifies baseline performance for a core in percent.

The `boot_disk` block supports:

* `auto_delete` - (Optional) Defines whether the disk will be auto-deleted when the instance
    is deleted. Default value is `True`.

* `device_name` - (Optional) Name that can be used to access an attached disk.

* `mode` - (Optional) Access mode of the Disk resource. By default, a disk is attached in `READ_WRITE` mode.

* `initialize_params` - (Optional) Parameters for a new disk that will be created
    alongside the new instance. Either `initialize_params` or `disk_id` must be set.
    The structure is documented below.

* `disk_id` - (Optional) The ID of the existing disk (such as those managed by
    `yandex_compute_disk`) to attach as a boot disk.

~> **NOTE:** Either `initialize_params` or `disk_id` must be specified.

The `initialize_params` block supports:

* `name` - (Optional) Name of the boot disk.

* `description` - (Optional) Description of the boot disk.

* `size` - (Optional) Size of the disk in GB.

* `type_id` - (Optional) Disk type.

* `image_id` - (Optional) A disk image to initialize this disk from.

* `snapshot_id` - (Optional) A snapshot to initialize this disk from.

~> **NOTE:** Either `image_id` or `snapshot_id` must be specified.

The `secondary_disk` block supports:

* `disk_id` - (Required) ID of the disk that is attached to the instance.

* `auto_delete` - (Optional) Whether the disk is auto-deleted when the instance
    is deleted. The default value is false.

* `device_name` - (Optional) Name that can be used to access an attached disk
    under `/dev/disk/by-id/`.

* `mode` - (Optional) Access mode to the Disk resource. By default, a disk is attached in `READ_WRITE` mode.

The `network_interface` block supports:

* `subnet_id` - (Optional) ID of the subnet to attach this
    interface to. The subnet must exist in the same zone where this instance will be
    created.

* `ip_address` - (Optional) The private IP address to assign to the instance. If
    empty, the address will be automatically assigned from the specified subnet.

* `ipv6` - (Optional) If true allocate IPv6 address for interaface.
    The address will be automatically assigned from the specified subnet.

* `ipv6_address` - (Optional) The private IPv6 address to assign to the instance.

* `nat` - (Optional) Provide a public address, for instance, to access the Internet through NAT.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `instance_id` - The server-assigned unique identifier of this instance.

* `network_interface.0.address` - The internal IP address of the instance.

* `network_interface.0.nat_ip_address` - The external IP address of the instance.

## Import

Instances can be imported using the `ID` of an instance, e.g.

```
$ terraform import yandex_compute_instance.default instance_id
```
