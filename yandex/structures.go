package yandex

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
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
		"gpus":          int(instance.Resources.Gpus),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func flattenInstanceGroupInstanceTemplateResources(resSpec *instancegroup.ResourcesSpec) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"cores":         int(resSpec.Cores),
		"core_fraction": int(resSpec.CoreFraction),
		"memory":        toGigabytesInFloat(resSpec.Memory),
		"gpus":          int(resSpec.Gpus),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func flattenInstanceBootDisk(ctx context.Context, instance *compute.Instance, diskServiceClient ReducedDiskServiceClient) ([]map[string]interface{}, error) {
	attachedDisk := instance.GetBootDisk()
	if attachedDisk == nil {
		return nil, nil
	}

	bootDisk := map[string]interface{}{
		"auto_delete": attachedDisk.GetAutoDelete(),
		"device_name": attachedDisk.GetDeviceName(),
		"disk_id":     attachedDisk.GetDiskId(),
		"mode":        attachedDisk.GetMode().String(),
	}

	disk, err := diskServiceClient.Get(ctx, &compute.GetDiskRequest{
		DiskId: attachedDisk.GetDiskId(),
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
			"mode":        instanceDisk.GetMode().String(),
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

func flattenInstanceGroupManagedInstanceNetworkInterfaces(instance *instancegroup.ManagedInstance) ([]map[string]interface{}, string, string, error) {
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

func flattenInstanceGroupInstanceTemplate(template *instancegroup.InstanceTemplate) ([]map[string]interface{}, error) {
	templateMap := make(map[string]interface{})

	templateMap["description"] = template.GetDescription()
	templateMap["labels"] = template.GetLabels()
	templateMap["platform_id"] = template.GetPlatformId()
	templateMap["metadata"] = template.GetMetadata()
	templateMap["service_account_id"] = template.GetServiceAccountId()

	resourceSpec, err := flattenInstanceGroupInstanceTemplateResources(template.GetResourcesSpec())
	if err != nil {
		return nil, err
	}
	templateMap["resources"] = resourceSpec

	bootDiskSpec, err := flattenInstanceGroupAttachedDisk(template.GetBootDiskSpec())
	if err != nil {
		return []map[string]interface{}{templateMap}, err
	}
	templateMap["boot_disk"] = []map[string]interface{}{bootDiskSpec}

	secondarySpecs := template.GetSecondaryDiskSpecs()
	secondarySpecsList := make([]map[string]interface{}, len(secondarySpecs))
	for i, spec := range secondarySpecs {
		flattened, err := flattenInstanceGroupAttachedDisk(spec)
		if err != nil {
			return nil, err
		}
		secondarySpecsList[i] = flattened
	}
	templateMap["secondary_disk"] = secondarySpecsList

	networkSpecs := template.GetNetworkInterfaceSpecs()
	networkSpecsList := make([]map[string]interface{}, len(networkSpecs))
	for i, spec := range networkSpecs {
		networkSpecsList[i] = flattenInstanceGroupNetworkInterfaceSpec(spec)
	}
	templateMap["network_interface"] = networkSpecsList

	if template.SchedulingPolicy != nil {
		templateMap["scheduling_policy"] = []map[string]interface{}{{"preemptible": template.SchedulingPolicy.Preemptible}}
	}

	return []map[string]interface{}{templateMap}, nil
}

func flattenInstanceGroupAttachedDisk(diskSpec *instancegroup.AttachedDiskSpec) (map[string]interface{}, error) {
	bootDisk := map[string]interface{}{
		"device_name": diskSpec.GetDeviceName(),
		"mode":        diskSpec.GetMode().String(),
	}

	diskSpecSpec := diskSpec.GetDiskSpec()

	if diskSpec == nil {
		return bootDisk, fmt.Errorf("no disk spec")
	}

	bootDisk["initialize_params"] = []map[string]interface{}{{
		"description": diskSpecSpec.Description,
		"size":        toGigabytes(diskSpecSpec.Size),
		"type":        diskSpecSpec.TypeId,
		"image_id":    diskSpecSpec.GetImageId(),
		"snapshot_id": diskSpecSpec.GetSnapshotId(),
	}}

	return bootDisk, nil
}

func flattenInstanceGroupNetworkInterfaceSpec(netSpec *instancegroup.NetworkInterfaceSpec) map[string]interface{} {
	nat := (netSpec.PrimaryV4AddressSpec != nil && netSpec.PrimaryV4AddressSpec.GetOneToOneNatSpec() != nil) ||
		(netSpec.PrimaryV6AddressSpec != nil && netSpec.PrimaryV6AddressSpec.GetOneToOneNatSpec() != nil)

	subnets := &schema.Set{F: schema.HashString}

	if netSpec.SubnetIds != nil {
		for _, s := range netSpec.SubnetIds {
			subnets.Add(s)
		}
	}

	networkInterface := map[string]interface{}{
		"network_id": netSpec.NetworkId,
		"subnet_ids": subnets,
		"nat":        nat,
		"ipv6":       netSpec.PrimaryV6AddressSpec != nil,
	}

	return networkInterface
}

func flattenInstanceGroupDeployPolicy(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}
	if ig.DeployPolicy != nil {
		res["max_expansion"] = ig.DeployPolicy.MaxExpansion
		res["max_creating"] = ig.DeployPolicy.MaxCreating
		res["max_deleting"] = ig.DeployPolicy.MaxDeleting
		res["max_unavailable"] = ig.DeployPolicy.MaxUnavailable
		if ig.DeployPolicy.StartupDuration != nil {
			res["startup_duration"] = ig.DeployPolicy.StartupDuration.Seconds
		}
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupScalePolicy(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	if sp := ig.GetScalePolicy().GetFixedScale(); sp != nil {
		res["fixed_scale"] = []map[string]interface{}{{"size": int(sp.Size)}}
		return []map[string]interface{}{res}, nil
	}

	if sp := ig.GetScalePolicy().GetAutoScale(); sp != nil {
		subres := map[string]interface{}{}
		res["auto_scale"] = []map[string]interface{}{subres}
		subres["min_zone_size"] = int(sp.MinZoneSize)
		subres["max_size"] = int(sp.MaxSize)
		subres["initial_size"] = int(sp.InitialSize)

		if sp.MeasurementDuration != nil {
			subres["measurement_duration"] = int(sp.MeasurementDuration.Seconds)
		}

		if sp.WarmupDuration != nil {
			subres["warmup_duration"] = int(sp.WarmupDuration.Seconds)
		}

		if sp.StabilizationDuration != nil {
			subres["stabilization_duration"] = int(sp.StabilizationDuration.Seconds)
		}

		if sp.CpuUtilizationRule != nil {
			subres["cpu_utilization_target"] = sp.CpuUtilizationRule.UtilizationTarget
		}

		if len(sp.CustomRules) > 0 {
			rules := make([]map[string]interface{}, len(sp.CustomRules))
			subres["custom_rule"] = rules

			for i, rule := range sp.CustomRules {
				rules[i] = map[string]interface{}{
					"rule_type":   instancegroup.ScalePolicy_CustomRule_RuleType_name[int32(rule.RuleType)],
					"metric_type": instancegroup.ScalePolicy_CustomRule_MetricType_name[int32(rule.MetricType)],
					"metric_name": rule.MetricName,
					"target":      rule.Target,
				}
			}
		}
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupAllocationPolicy(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	zones := &schema.Set{F: schema.HashString}

	for _, zone := range ig.AllocationPolicy.Zones {
		zones.Add(zone.ZoneId)
	}

	res["zones"] = zones
	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupHealthChecks(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	if ig.HealthChecksSpec == nil {
		return nil, nil
	}

	res := make([]map[string]interface{}, len(ig.HealthChecksSpec.HealthCheckSpecs))

	for i, spec := range ig.HealthChecksSpec.HealthCheckSpecs {
		specDict := map[string]interface{}{}
		specDict["interval"] = int(spec.Interval.Seconds)
		specDict["timeout"] = int(spec.Timeout.Seconds)
		specDict["healthy_threshold"] = int(spec.HealthyThreshold)
		specDict["unhealthy_threshold"] = int(spec.UnhealthyThreshold)

		if spec.GetHttpOptions() != nil {
			specDict["http_options"] = []map[string]interface{}{
				{
					"port": int(spec.GetHttpOptions().Port),
					"path": spec.GetHttpOptions().Path,
				},
			}
		}

		if spec.GetTcpOptions() != nil {
			specDict["tcp_options"] = []map[string]interface{}{
				{
					"port": int(spec.GetTcpOptions().Port),
				},
			}
		}

		res[i] = specDict
	}
	return res, nil
}

func flattenInstanceGroupLoadBalancerState(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	if loadBalancerState := ig.GetLoadBalancerState(); loadBalancerState != nil {
		res["target_group_id"] = loadBalancerState.TargetGroupId
		res["status_message"] = loadBalancerState.StatusMessage
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupLoadBalancerSpec(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	if ig.LoadBalancerSpec == nil || ig.LoadBalancerSpec.TargetGroupSpec == nil {
		return nil, nil
	}

	res := map[string]interface{}{}
	res["target_group_name"] = ig.LoadBalancerSpec.TargetGroupSpec.GetName()
	res["target_group_description"] = ig.LoadBalancerSpec.TargetGroupSpec.GetDescription()
	res["target_group_labels"] = ig.LoadBalancerSpec.TargetGroupSpec.GetLabels()
	res["target_group_id"] = ig.LoadBalancerState.GetTargetGroupId()
	res["status_message"] = ig.LoadBalancerState.GetStatusMessage()

	return []map[string]interface{}{res}, nil
}

func expandInstanceResourcesSpec(d *schema.ResourceData) (*compute.ResourcesSpec, error) {
	rs := &compute.ResourcesSpec{}

	if v, ok := d.GetOk("resources.0.cores"); ok {
		rs.Cores = int64(v.(int))
	}

	if v, ok := d.GetOk("resources.0.gpus"); ok {
		rs.Gpus = int64(v.(int))
	}

	if v, ok := d.GetOk("resources.0.core_fraction"); ok {
		rs.CoreFraction = int64(v.(int))
	}

	if v, ok := d.GetOk("resources.0.memory"); ok {
		rs.Memory = toBytesFromFloat(v.(float64))
	}

	return rs, nil
}

func expandInstanceGroupResourcesSpec(d *schema.ResourceData, prefix string) (*instancegroup.ResourcesSpec, error) {
	rs := &instancegroup.ResourcesSpec{}

	if v, ok := d.GetOk(prefix + ".0.cores"); ok {
		rs.Cores = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + ".0.gpus"); ok {
		rs.Gpus = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + ".0.core_fraction"); ok {
		rs.CoreFraction = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + ".0.memory"); ok {
		rs.Memory = toBytesFromFloat(v.(float64))
	}

	return rs, nil
}

func expandInstanceBootDiskSpec(d *schema.ResourceData, config *Config) (*compute.AttachedDiskSpec, error) {
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
		bootDiskSpec, err := expandBootDiskSpec(d, config)
		if err != nil {
			return nil, err
		}

		ads.Disk = &compute.AttachedDiskSpec_DiskSpec_{
			DiskSpec: bootDiskSpec,
		}
	}

	return ads, nil
}

func expandInstanceGroupTemplateAttachedDiskSpec(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.AttachedDiskSpec, error) {
	ads := &instancegroup.AttachedDiskSpec{}

	if v, ok := d.GetOk(prefix + ".device_name"); ok {
		ads.DeviceName = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".mode"); ok {
		diskMode, err := parseInstanceGroupDiskMode(v.(string))
		if err != nil {
			return nil, err
		}

		ads.Mode = diskMode
	}

	// create new one disk
	if _, ok := d.GetOk(prefix + ".initialize_params"); ok {
		bootDiskSpec, err := expandInstanceGroupAttachenDiskSpecSpec(d, prefix+".initialize_params.0", config)
		if err != nil {
			return nil, err
		}

		ads.DiskSpec = bootDiskSpec
	}

	return ads, nil
}

func expandBootDiskSpec(d *schema.ResourceData, config *Config) (*compute.AttachedDiskSpec_DiskSpec, error) {
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

	var minStorageSizeBytes int64
	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.image_id"); ok {
		imageID := v.(string)
		diskSpec.Source = &compute.AttachedDiskSpec_DiskSpec_ImageId{
			ImageId: imageID,
		}

		size, err := getImageMinStorageSize(imageID, config)
		if err != nil {
			return nil, err
		}
		minStorageSizeBytes = size
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.snapshot_id"); ok {
		snapshotID := v.(string)
		diskSpec.Source = &compute.AttachedDiskSpec_DiskSpec_SnapshotId{
			SnapshotId: snapshotID,
		}

		size, err := getSnapshotMinStorageSize(snapshotID, config)
		if err != nil {
			return nil, err
		}
		minStorageSizeBytes = size
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.size"); ok {
		diskSpec.Size = toBytes(v.(int))
	}

	if diskSpec.Size == 0 {
		diskSpec.Size = minStorageSizeBytes
	}

	return diskSpec, nil
}

func expandInstanceGroupAttachenDiskSpecSpec(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.AttachedDiskSpec_DiskSpec, error) {
	diskSpec := &instancegroup.AttachedDiskSpec_DiskSpec{}

	if v, ok := d.GetOk(prefix + ".description"); ok {
		diskSpec.Description = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".type"); ok {
		diskSpec.TypeId = v.(string)
	}

	var minStorageSizeBytes int64
	if v, ok := d.GetOk(prefix + ".image_id"); ok {
		imageID := v.(string)
		diskSpec.SourceOneof = &instancegroup.AttachedDiskSpec_DiskSpec_ImageId{
			ImageId: imageID,
		}

		size, err := getImageMinStorageSize(imageID, config)
		if err != nil {
			return nil, err
		}
		minStorageSizeBytes = size
	}

	if v, ok := d.GetOk(prefix + ".snapshot_id"); ok {
		snapshotID := v.(string)
		diskSpec.SourceOneof = &instancegroup.AttachedDiskSpec_DiskSpec_SnapshotId{
			SnapshotId: snapshotID,
		}

		size, err := getSnapshotMinStorageSize(snapshotID, config)
		if err != nil {
			return nil, err
		}
		minStorageSizeBytes = size
	}

	if v, ok := d.GetOk(prefix + ".size"); ok {
		diskSpec.Size = toBytes(v.(int))
	}

	if diskSpec.Size == 0 {
		diskSpec.Size = minStorageSizeBytes
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

func expandInstanceGroupSecondaryDiskSpecs(d *schema.ResourceData, prefix string, config *Config) ([]*instancegroup.AttachedDiskSpec, error) {
	secondaryDisksCount := d.Get(prefix + ".#").(int)
	ads := make([]*instancegroup.AttachedDiskSpec, secondaryDisksCount)

	for i := 0; i < secondaryDisksCount; i++ {
		disk, err := expandInstanceGroupTemplateAttachedDiskSpec(d, fmt.Sprintf(prefix+".%d", i), config)
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

		if nat, ok := data["nat"].(bool); ok && nat {
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

func expandInstanceGroupNetworkInterfaceSpecs(d *schema.ResourceData, prefix string) ([]*instancegroup.NetworkInterfaceSpec, error) {
	nicsConfig := d.Get(prefix).([]interface{})
	nics := make([]*instancegroup.NetworkInterfaceSpec, len(nicsConfig))

	for i, raw := range nicsConfig {
		data := raw.(map[string]interface{})

		nics[i] = &instancegroup.NetworkInterfaceSpec{
			NetworkId: data["network_id"].(string),
		}

		if subnets, ok := data["subnet_ids"]; ok {
			sub := subnets.(*schema.Set).List()

			nics[i].SubnetIds = make([]string, 0)

			for _, s := range sub {
				nics[i].SubnetIds = append(nics[i].SubnetIds, s.(string))
			}
		}

		if enableIPV6, ok := data["ipv6"].(bool); ok && enableIPV6 {
			nics[i].PrimaryV6AddressSpec = &instancegroup.PrimaryAddressSpec{}
		}

		if nat, ok := data["nat"].(bool); ok && nat {
			nics[i].PrimaryV4AddressSpec = &instancegroup.PrimaryAddressSpec{
				OneToOneNatSpec: &instancegroup.OneToOneNatSpec{
					IpVersion: instancegroup.IpVersion_IPV4,
				},
			}

		} else {
			nics[i].PrimaryV4AddressSpec = &instancegroup.PrimaryAddressSpec{}
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

func parseInstanceGroupDiskMode(mode string) (instancegroup.AttachedDiskSpec_Mode, error) {
	val, ok := compute.AttachedDiskSpec_Mode_value[mode]
	if !ok {
		return instancegroup.AttachedDiskSpec_MODE_UNSPECIFIED, fmt.Errorf("value for 'mode' should be 'READ_WRITE' or 'READ_ONLY', not '%s'", mode)
	}
	return instancegroup.AttachedDiskSpec_Mode(val), nil
}

func parseIamKeyAlgorithm(algorithm string) (iam.Key_Algorithm, error) {
	val, ok := iam.Key_Algorithm_value[algorithm]
	if !ok {
		return iam.Key_ALGORITHM_UNSPECIFIED, fmt.Errorf("value for 'key_algorithm' should be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(iam.KeyFormat_value)), algorithm)
	}
	return iam.Key_Algorithm(val), nil
}

func parseIamKeyFormat(format string) (iam.KeyFormat, error) {
	val, ok := iam.KeyFormat_value[format]
	if !ok {
		return iam.KeyFormat(0), fmt.Errorf("value for 'format' should be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(iam.KeyFormat_value)), format)
	}
	return iam.KeyFormat(val), nil
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

// revive:disable:var-naming
func expandInstanceGroupInstanceTemplate(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.InstanceTemplate, error) {
	var platformId, description, serviceAccount string

	if v, ok := d.GetOk(prefix + ".platform_id"); ok {
		platformId = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".description"); ok {
		description = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".service_account_id"); ok {
		serviceAccount = v.(string)
	}

	resourceSpec, err := expandInstanceGroupResourcesSpec(d, prefix+".resources")
	if err != nil {
		return nil, fmt.Errorf("Error create 'resources' object of api request: %s", err)
	}

	bootDiskSpec, err := expandInstanceGroupTemplateAttachedDiskSpec(d, prefix+".boot_disk.0", config)
	if err != nil {
		return nil, fmt.Errorf("Error create 'boot_disk' object of api request: %s", err)
	}

	secondaryDiskSpecs, err := expandInstanceGroupSecondaryDiskSpecs(d, prefix+".secondary_disk", config)
	if err != nil {
		return nil, fmt.Errorf("Error create 'secondary_disk' object of api request: %s", err)
	}

	nicSpecs, err := expandInstanceGroupNetworkInterfaceSpecs(d, prefix+".network_interface")
	if err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	schedulingPolicy, err := expandInstanceGroupSchedulingPolicy(d, prefix+".scheduling_policy")
	if err != nil {
		return nil, fmt.Errorf("Error create 'scheduling_policy' object of api request: %s", err)
	}

	labels, err := expandLabels(d.Get(prefix + ".labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance group: %s", err)
	}

	metadata, err := expandLabels(d.Get(prefix + ".metadata"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding metadata while creating instance group: %s", err)
	}

	template := &instancegroup.InstanceTemplate{
		BootDiskSpec:          bootDiskSpec,
		Description:           description,
		Labels:                labels,
		Metadata:              metadata,
		NetworkInterfaceSpecs: nicSpecs,
		PlatformId:            platformId,
		ResourcesSpec:         resourceSpec,
		SchedulingPolicy:      schedulingPolicy,
		SecondaryDiskSpecs:    secondaryDiskSpecs,
		ServiceAccountId:      serviceAccount,
	}

	return template, nil
}

func expandInstanceGroupScalePolicy(d *schema.ResourceData) (*instancegroup.ScalePolicy, error) {
	if _, ok := d.GetOk("scale_policy.0.fixed_scale"); ok {
		v := d.Get("scale_policy.0.fixed_scale.0.size").(int)
		policy := &instancegroup.ScalePolicy{
			ScaleType: &instancegroup.ScalePolicy_FixedScale_{
				FixedScale: &instancegroup.ScalePolicy_FixedScale{Size: int64(v)},
			}}
		return policy, nil
	}

	if _, ok := d.GetOk("scale_policy.0.auto_scale"); ok {
		autoScale := &instancegroup.ScalePolicy_AutoScale{
			MinZoneSize: int64(d.Get("scale_policy.0.auto_scale.0.min_zone_size").(int)),
			MaxSize:     int64(d.Get("scale_policy.0.auto_scale.0.max_size").(int)),
			InitialSize: int64(d.Get("scale_policy.0.auto_scale.0.initial_size").(int)),
		}

		if v, ok := d.GetOk("scale_policy.0.auto_scale.0.measurement_duration"); ok {
			autoScale.MeasurementDuration = &duration.Duration{Seconds: int64(v.(int))}
		}

		if v, ok := d.GetOk("scale_policy.0.auto_scale.0.warmup_duration"); ok {
			autoScale.WarmupDuration = &duration.Duration{Seconds: int64(v.(int))}
		}

		if v, ok := d.GetOk("scale_policy.0.auto_scale.0.cpu_utilization_target"); ok {
			autoScale.CpuUtilizationRule = &instancegroup.ScalePolicy_CpuUtilizationRule{UtilizationTarget: v.(float64)}
		}

		if v, ok := d.GetOk("scale_policy.0.auto_scale.0.stabilization_duration"); ok {
			autoScale.StabilizationDuration = &duration.Duration{Seconds: int64(v.(int))}
		}

		if customRulesCount := d.Get("scale_policy.0.auto_scale.0.custom_rule.#").(int); customRulesCount > 0 {
			rules := make([]*instancegroup.ScalePolicy_CustomRule, customRulesCount)
			for i := 0; i < customRulesCount; i++ {
				key := fmt.Sprintf("scale_policy.0.auto_scale.0.custom_rule.%d", i)
				if rule, err := expandCustomRule(d, key); err == nil {
					rules[i] = rule
				} else {
					return nil, err
				}
			}
			autoScale.CustomRules = rules
		}

		policy := &instancegroup.ScalePolicy{
			ScaleType: &instancegroup.ScalePolicy_AutoScale_{
				AutoScale: autoScale,
			}}
		return policy, nil
	}

	return nil, fmt.Errorf("Only fixed_scale and auto_scale policy are supported")
}

func expandCustomRule(d *schema.ResourceData, prefix string) (*instancegroup.ScalePolicy_CustomRule, error) {
	ruleType, ok := instancegroup.ScalePolicy_CustomRule_RuleType_value[d.Get(prefix+".rule_type").(string)]
	if !ok {
		return nil, fmt.Errorf("invalid value for rule_type")
	}

	metricType, ok := instancegroup.ScalePolicy_CustomRule_MetricType_value[d.Get(prefix+".metric_type").(string)]
	if !ok {
		return nil, fmt.Errorf("invalid value for metric_type")
	}

	return &instancegroup.ScalePolicy_CustomRule{
		RuleType:   instancegroup.ScalePolicy_CustomRule_RuleType(ruleType),
		MetricType: instancegroup.ScalePolicy_CustomRule_MetricType(metricType),
		MetricName: d.Get(prefix + ".metric_name").(string),
		Target:     d.Get(prefix + ".target").(float64),
	}, nil
}

func expandInstanceGroupDeployPolicy(d *schema.ResourceData) (*instancegroup.DeployPolicy, error) {
	policy := &instancegroup.DeployPolicy{
		MaxUnavailable: int64(d.Get("deploy_policy.0.max_unavailable").(int)),
		MaxDeleting:    int64(d.Get("deploy_policy.0.max_deleting").(int)),
		MaxCreating:    int64(d.Get("deploy_policy.0.max_creating").(int)),
		MaxExpansion:   int64(d.Get("deploy_policy.0.max_expansion").(int)),
	}

	if v, ok := d.GetOk("deploy_policy.0.startup_duration"); ok {
		policy.StartupDuration = &duration.Duration{Seconds: int64(v.(int))}
	}
	return policy, nil
}

func expandInstanceGroupAllocationPolicy(d *schema.ResourceData) (*instancegroup.AllocationPolicy, error) {
	if v, ok := d.GetOk("allocation_policy.0.zones"); ok {
		zones := make([]*instancegroup.AllocationPolicy_Zone, 0)

		for _, s := range v.(*schema.Set).List() {
			zones = append(zones, &instancegroup.AllocationPolicy_Zone{ZoneId: s.(string)})
		}

		policy := &instancegroup.AllocationPolicy{Zones: zones}
		return policy, nil
	}

	return nil, fmt.Errorf("Zones must be defined")
}

func expandInstanceGroupHealthCheckSpec(d *schema.ResourceData) (*instancegroup.HealthChecksSpec, error) {
	checksCount := d.Get("health_check.#").(int)

	if checksCount == 0 {
		return nil, nil
	}

	checks := make([]*instancegroup.HealthCheckSpec, checksCount)

	for i := 0; i < checksCount; i++ {
		key := fmt.Sprintf("health_check.%d", i)
		hc := &instancegroup.HealthCheckSpec{
			HealthyThreshold:   int64(d.Get(key + ".healthy_threshold").(int)),
			UnhealthyThreshold: int64(d.Get(key + ".unhealthy_threshold").(int)),
		}
		if v, ok := d.GetOk(key + ".interval"); ok {
			hc.Interval = &duration.Duration{Seconds: int64(v.(int))}
		}
		if v, ok := d.GetOk(key + ".timeout"); ok {
			hc.Timeout = &duration.Duration{Seconds: int64(v.(int))}
		}
		checks[i] = hc

		if _, ok := d.GetOk(key + ".tcp_options"); ok {
			hc.HealthCheckOptions = &instancegroup.HealthCheckSpec_TcpOptions_{TcpOptions: &instancegroup.HealthCheckSpec_TcpOptions{Port: int64(d.Get(key + ".tcp_options.0.port").(int))}}
			continue
		}

		if _, ok := d.GetOk(key + ".http_options"); ok {
			hc.HealthCheckOptions = &instancegroup.HealthCheckSpec_HttpOptions_{
				HttpOptions: &instancegroup.HealthCheckSpec_HttpOptions{Port: int64(d.Get(key + ".http_options.0.port").(int)), Path: d.Get(key + ".http_options.0.path").(string)},
			}
			continue
		}

		return nil, fmt.Errorf("need tcp_options or http_options")
	}

	return &instancegroup.HealthChecksSpec{HealthCheckSpecs: checks}, nil
}

func expandInstanceGroupLoadBalancerSpec(d *schema.ResourceData) (*instancegroup.LoadBalancerSpec, error) {
	if _, ok := d.GetOk("load_balancer"); !ok {
		return nil, nil
	}

	spec := &instancegroup.TargetGroupSpec{
		Name:        d.Get("load_balancer.0.target_group_name").(string),
		Description: d.Get("load_balancer.0.target_group_description").(string),
	}

	if v, ok := d.GetOk("load_balancer.0.target_group_labels"); ok {
		labels, err := expandLabels(v)
		if err != nil {
			return nil, fmt.Errorf("Error expanding labels: %s", err)
		}

		spec.Labels = labels
	}

	return &instancegroup.LoadBalancerSpec{TargetGroupSpec: spec}, nil
}

func expandInstanceGroupSchedulingPolicy(d *schema.ResourceData, prefix string) (*instancegroup.SchedulingPolicy, error) {
	p := d.Get(prefix + ".0.preemptible").(bool)
	return &instancegroup.SchedulingPolicy{Preemptible: p}, nil
}

func flattenInstances(instances []*instancegroup.ManagedInstance) ([]map[string]interface{}, error) {
	if instances == nil {
		return []map[string]interface{}{}, nil
	}

	res := make([]map[string]interface{}, len(instances))

	for i, instance := range instances {
		instDict := make(map[string]interface{})
		instDict["status"] = instance.GetStatus().String()
		instDict["instance_id"] = instance.GetInstanceId()
		instDict["fqdn"] = instance.GetFqdn()
		instDict["name"] = instance.GetName()
		instDict["status_message"] = instance.GetStatusMessage()
		instDict["zone_id"] = instance.GetZoneId()

		networkInterfaces, _, _, err := flattenInstanceGroupManagedInstanceNetworkInterfaces(instance)
		if err != nil {
			return res, err
		}

		instDict["network_interface"] = networkInterfaces
		res[i] = instDict
	}

	return res, nil
}
