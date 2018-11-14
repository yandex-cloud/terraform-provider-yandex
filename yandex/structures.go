package yandex

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

type ReducedDiskServiceClient interface {
	Get(ctx context.Context, in *compute.GetDiskRequest, opts ...grpc.CallOption) (*compute.Disk, error)
}

func expandLabels(v interface{}) (map[string]string, error) {
	m := make(map[string]string)
	if v == nil {
		return m, nil
	}
	for k, val := range v.(map[string]interface{}) {
		m[k] = val.(string)
	}
	return m, nil
}

func expandProductIds(v interface{}) ([]string, error) {
	m := []string{}
	if v == nil {
		return m, nil
	}
	tagsSet := v.(*schema.Set)
	for _, val := range tagsSet.List() {
		m = append(m, val.(string))
	}
	return m, nil
}

func flattenInstanceResources(instance *compute.Instance) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"cores":         int(instance.Resources.Cores),
		"core_fraction": int(instance.Resources.CoreFraction),
		"memory":        toGigabytes(instance.Resources.Memory),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func flattenInstanceBootDisk(instance *compute.Instance, diskServiceClient ReducedDiskServiceClient) ([]map[string]interface{}, error) {
	bootDisk := map[string]interface{}{
		"auto_delete": instance.BootDisk.AutoDelete,
		"device_name": instance.BootDisk.DeviceName,
		"disk_id":     instance.BootDisk.DiskId,
		"mode":        instance.BootDisk.Mode.String(),
	}

	disk, err := diskServiceClient.Get(context.Background(), &compute.GetDiskRequest{
		DiskId: instance.BootDisk.DiskId,
	})
	if err != nil {
		return nil, err
	}

	bootDisk["initialize_params"] = []map[string]interface{}{{
		"name":        disk.Name,
		"description": disk.Description,
		"size":        toGigabytes(disk.Size),
		"type_id":     disk.TypeId,
		"image_id":    disk.GetSourceImageId(),
		"snapshot_id": disk.GetSourceSnapshotId(),
	}}

	return []map[string]interface{}{bootDisk}, nil
}

func flattenInstanceSecondaryDisks(instance *compute.Instance) ([]map[string]interface{}, error) {
	secondaryDisks := []map[string]interface{}{}

	for _, instanceDisk := range instance.SecondaryDisks {
		disk := map[string]interface{}{
			"disk_id":     instanceDisk.DiskId,
			"device_name": instanceDisk.DeviceName,
			"mode":        instanceDisk.Mode.String(),
			"auto_delete": instanceDisk.AutoDelete,
		}
		secondaryDisks = append(secondaryDisks, disk)
	}
	return secondaryDisks, nil
}

func flattenInstanceNetworkInterfaces(instance *compute.Instance) ([]map[string]interface{}, string, error) {
	nics := make([]map[string]interface{}, len(instance.NetworkInterfaces))
	var externalIP string

	for i, iface := range instance.NetworkInterfaces {
		index, err := strconv.Atoi(iface.Index)
		if err != nil {
			return nil, "", fmt.Errorf("Error while convert index of Network Interface: %s", err)
		}

		nics[i] = map[string]interface{}{
			"index":       index,
			"mac_address": iface.MacAddress,
			"subnet_id":   iface.SubnetId,
		}

		if iface.PrimaryV4Address != nil {
			nics[i]["ip_address"] = iface.PrimaryV4Address.Address

			if iface.PrimaryV4Address.OneToOneNat != nil {
				nics[i]["nat"] = true
				nics[i]["nat_ip_address"] = iface.PrimaryV4Address.OneToOneNat.Address
				nics[i]["nat_ip_version"] = iface.PrimaryV4Address.OneToOneNat.IpVersion.String()
				if externalIP == "" {
					externalIP = iface.PrimaryV4Address.OneToOneNat.Address
				}
			} else {
				nics[i]["nat"] = false
			}
		}

		if iface.PrimaryV6Address != nil {
			nics[i]["ipv6"] = true
			nics[i]["ipv6_address"] = iface.PrimaryV6Address.Address
		}
	}

	return nics, externalIP, nil
}

func expandInstanceResourcesSpec(d *schema.ResourceData) (*compute.ResourcesSpec, error) {
	rs := &compute.ResourcesSpec{}

	if v, ok := d.GetOk("resources"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			res := v.(map[string]interface{})

			if v, ok := res["cores"].(int); ok {
				rs.Cores = int64(v)
			}

			if v, ok := res["core_fraction"].(int); ok {
				rs.CoreFraction = int64(v)
			}

			if v, ok := res["memory"].(int); ok && v != 0 {
				rs.Memory = toBytes(v)
			}
		}
	} else {
		// should not occur: validation must be done at Schema level
		return nil, fmt.Errorf("You should define 'resources' section for compute instance")
	}

	return rs, nil
}

func expandInstanceBootDiskSpec(d *schema.ResourceData) (*compute.AttachedDiskSpec, error) {
	ads := &compute.AttachedDiskSpec{}

	if v, ok := d.GetOk("boot_disk.0.auto_delete"); ok {
		ads.AutoDelete = v.(bool)
	}

	if v, ok := d.GetOk("boot_disk.0.device_name"); ok {
		ads.DeviceName = v.(string)
	}

	if v, ok := d.GetOk("boot_disk.0.mode"); ok {
		diskMode, err := parseDiskMode(v.(string))
		if err != nil {
			return nil, err
		}

		ads.Mode = diskMode
	}

	// use explicit disk
	if v, ok := d.GetOk("boot_disk.0.disk_id"); ok {
		ads.Disk = &compute.AttachedDiskSpec_DiskId{
			DiskId: v.(string),
		}
		return ads, nil
	}

	// create new one disk
	if _, ok := d.GetOk("boot_disk.0.initialize_params"); ok {
		bootDiskSpec, err := expandBootDiskSpec(d)
		if err != nil {
			return nil, err
		}

		ads.Disk = &compute.AttachedDiskSpec_DiskSpec_{
			DiskSpec: bootDiskSpec,
		}
	}

	return ads, nil
}

func expandBootDiskSpec(d *schema.ResourceData) (*compute.AttachedDiskSpec_DiskSpec, error) {
	diskSpec := &compute.AttachedDiskSpec_DiskSpec{}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.name"); ok {
		diskSpec.Name = v.(string)
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.description"); ok {
		diskSpec.Description = v.(string)
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.type_id"); ok {
		diskSpec.TypeId = v.(string)
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.size"); ok {
		diskSpec.Size = toBytes(v.(int))
	}

	if v, b := d.GetOk("boot_disk.0.initialize_params.0.image_id"); b {
		diskSpec.Source = &compute.AttachedDiskSpec_DiskSpec_ImageId{
			ImageId: v.(string),
		}
	}

	if v, b := d.GetOk("boot_disk.0.initialize_params.0.snapshot_id"); b {
		diskSpec.Source = &compute.AttachedDiskSpec_DiskSpec_SnapshotId{
			SnapshotId: v.(string),
		}
	}

	return diskSpec, nil
}

func expandInstanceSecondaryDiskSpecs(d *schema.ResourceData) ([]*compute.AttachedDiskSpec, error) {
	secondaryDisksCount := d.Get("secondary_disk.#").(int)
	ads := make([]*compute.AttachedDiskSpec, secondaryDisksCount)

	for i := 0; i < secondaryDisksCount; i++ {
		diskConfig := d.Get(fmt.Sprintf("secondary_disk.%d", i)).(map[string]interface{})

		disk, err := expandSecondaryDiskSpec(diskConfig)
		if err != nil {
			return nil, err
		}
		ads[i] = disk
	}
	return ads, nil
}

func expandSecondaryDiskSpec(diskConfig map[string]interface{}) (*compute.AttachedDiskSpec, error) {
	disk := &compute.AttachedDiskSpec{}

	if v, ok := diskConfig["mode"]; ok {
		mode, err := parseDiskMode(v.(string))
		if err != nil {
			return nil, err
		}
		disk.Mode = mode
	}

	if v, ok := diskConfig["device_name"]; ok {
		disk.DeviceName = v.(string)
	}

	if v, ok := diskConfig["auto_delete"]; ok {
		disk.AutoDelete = v.(bool)
	}

	if v, ok := diskConfig["disk_id"]; ok {
		// TODO: support disk creation
		disk.Disk = &compute.AttachedDiskSpec_DiskId{
			DiskId: v.(string),
		}
	}

	return disk, nil
}

func expandInstanceNetworkInterfaceSpecs(d *schema.ResourceData) ([]*compute.NetworkInterfaceSpec, error) {
	nicsConfig := d.Get("network_interface").([]interface{})
	nics := make([]*compute.NetworkInterfaceSpec, len(nicsConfig))

	for i, raw := range nicsConfig {
		data := raw.(map[string]interface{})

		subnetID := data["subnet_id"].(string)
		if subnetID == "" {
			return nil, fmt.Errorf("NIC number %d does not have a 'subnet_id' attribute defined", i)
		}

		nics[i] = &compute.NetworkInterfaceSpec{
			SubnetId: subnetID,
		}

		ipV4Address := data["ip_address"].(string)
		ipV6Address := data["ipv6_address"].(string)

		// By default allocate any unassigned IPv4 address
		if ipV4Address == "" && ipV6Address == "" {
			nics[i].PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{}
		}

		if enableIPV6, ok := data["ipv6"].(bool); ok && enableIPV6 {
			nics[i].PrimaryV6AddressSpec = &compute.PrimaryAddressSpec{}
		}

		if ipV4Address != "" {
			nics[i].PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{
				Address: ipV4Address,
			}
		}

		if ipV6Address != "" {
			nics[i].PrimaryV6AddressSpec = &compute.PrimaryAddressSpec{
				Address: ipV6Address,
			}
		}

		if enableNat, ok := data["nat"].(bool); ok && enableNat {
			if nics[i].PrimaryV4AddressSpec == nil {
				nics[i].PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{
					OneToOneNatSpec: &compute.OneToOneNatSpec{
						IpVersion: compute.IpVersion_IPV4,
					},
				}
			} else {
				nics[i].PrimaryV4AddressSpec.OneToOneNatSpec = &compute.OneToOneNatSpec{
					IpVersion: compute.IpVersion_IPV4,
				}
			}
		}
	}

	return nics, nil
}

func parseDiskMode(mode string) (compute.AttachedDiskSpec_Mode, error) {
	val, ok := compute.AttachedDiskSpec_Mode_value[mode]
	if !ok {
		return compute.AttachedDiskSpec_MODE_UNSPECIFIED, fmt.Errorf("value for 'mode' should be 'READ_WRITE' or 'READ_ONLY', not '%s'", mode)
	}
	return compute.AttachedDiskSpec_Mode(val), nil
}
