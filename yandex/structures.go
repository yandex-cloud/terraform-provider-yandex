package yandex

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
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
		"memory":        toGigabytesInFloat(instance.Resources.Memory),
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
		"type":        disk.TypeId,
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

func flattenInstanceNetworkInterfaces(instance *compute.Instance) ([]map[string]interface{}, string, string, error) {
	nics := make([]map[string]interface{}, len(instance.NetworkInterfaces))
	var externalIP, internalIP string

	for i, iface := range instance.NetworkInterfaces {
		index, err := strconv.Atoi(iface.Index)
		if err != nil {
			return nil, "", "", fmt.Errorf("Error while convert index of Network Interface: %s", err)
		}

		nics[i] = map[string]interface{}{
			"index":       index,
			"mac_address": iface.MacAddress,
			"subnet_id":   iface.SubnetId,
		}

		if iface.PrimaryV4Address != nil {
			nics[i]["ip_address"] = iface.PrimaryV4Address.Address
			if internalIP == "" {
				internalIP = iface.PrimaryV4Address.Address
			}

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
			if externalIP == "" {
				externalIP = iface.PrimaryV6Address.Address
			}
		}
	}

	return nics, externalIP, internalIP, nil
}

func expandInstanceResourcesSpec(d *schema.ResourceData) (*compute.ResourcesSpec, error) {
	rs := &compute.ResourcesSpec{}

	if v, ok := d.GetOk("resources.0.cores"); ok {
		rs.Cores = int64(v.(int))
	}

	if v, ok := d.GetOk("resources.0.core_fraction"); ok {
		rs.CoreFraction = int64(v.(int))
	}

	if v, ok := d.GetOk("resources.0.memory"); ok {
		rs.Memory = toBytesFromFloat(v.(float64))
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

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.type"); ok {
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

func expandInstanceSchedulingPolicy(d *schema.ResourceData) (*compute.SchedulingPolicy, error) {
	sp := d.Get("scheduling_policy").([]interface{})
	var schedulingPolicy *compute.SchedulingPolicy
	if len(sp) != 0 {
		schedulingPolicy = &compute.SchedulingPolicy{
			Preemptible: d.Get("scheduling_policy.0.preemptible").(bool),
		}
	}
	return schedulingPolicy, nil
}

func flattenInstanceSchedulingPolicy(instance *compute.Instance) ([]map[string]interface{}, error) {
	schedulingPolicy := make([]map[string]interface{}, 0, 1)
	schedulingMap := map[string]interface{}{
		"preemptible": instance.SchedulingPolicy.Preemptible,
	}
	schedulingPolicy = append(schedulingPolicy, schedulingMap)
	return schedulingPolicy, nil
}

func flattenStaticRoutes(routeTable *vpc.RouteTable) *schema.Set {
	staticRoutes := schema.NewSet(resourceYandexVPCRouteTableHash, nil)

	for _, r := range routeTable.StaticRoutes {
		m := make(map[string]interface{})

		switch d := r.Destination.(type) {
		case *vpc.StaticRoute_DestinationPrefix:
			m["destination_prefix"] = d.DestinationPrefix
		}

		switch h := r.NextHop.(type) {
		case *vpc.StaticRoute_NextHopAddress:
			m["next_hop_address"] = h.NextHopAddress
		}

		staticRoutes.Add(m)
	}
	return staticRoutes
}

func expandStaticRoutes(d *schema.ResourceData) ([]*vpc.StaticRoute, error) {
	staticRoutes := []*vpc.StaticRoute{}

	if v, ok := d.GetOk("static_route"); ok {
		routeList := v.(*schema.Set).List()
		for _, v := range routeList {
			sr, err := routeDescriptionToStaticRoute(v)
			if err != nil {
				return nil, fmt.Errorf("fail convert static route: %s", err)
			}
			staticRoutes = append(staticRoutes, sr)
		}
	} else {
		// should not occur: validation must be done at Schema level
		return nil, fmt.Errorf("You should define 'static_route' section for route table")
	}

	return staticRoutes, nil
}

func routeDescriptionToStaticRoute(v interface{}) (*vpc.StaticRoute, error) {
	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", v)
	}

	var sr vpc.StaticRoute

	if v, ok := res["destination_prefix"].(string); ok {
		sr.Destination = &vpc.StaticRoute_DestinationPrefix{
			DestinationPrefix: v,
		}
	}

	if v, ok := res["next_hop_address"].(string); ok {
		sr.NextHop = &vpc.StaticRoute_NextHopAddress{
			NextHopAddress: v,
		}
	}

	return &sr, nil
}
