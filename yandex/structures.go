package yandex

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
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
			"ipv4":        false,
			"ipv6":        false,
		}

		if iface.GetSecurityGroupIds() != nil {
			nics[i]["security_group_ids"] = convertStringArrToInterface(iface.SecurityGroupIds)
		}

		if iface.PrimaryV4Address != nil {
			nics[i]["ipv4"] = true
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
			"ipv4":        false,
			"ipv6":        false,
		}

		if iface.PrimaryV4Address != nil {
			nics[i]["ipv4"] = true
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
	templateMap["name"] = template.GetName()
	templateMap["hostname"] = template.GetHostname()

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

	if template.NetworkSettings != nil {
		templateMap["network_settings"] = flattenInstanceGroupNetworkSettings(template.GetNetworkSettings())
	}

	return []map[string]interface{}{templateMap}, nil
}

func flattenInstanceGroupVariable(v []*instancegroup.Variable) map[string]string {
	variables := make(map[string]string)
	for _, raw := range v {
		variables[raw.GetKey()] = raw.GetValue()
	}
	return variables
}

func flattenInstanceGroupNetworkSettings(ns *instancegroup.NetworkSettings) []map[string]interface{} {
	return []map[string]interface{}{{"type": ns.GetType().String()}}
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

func flattenInstanceGroupNetworkInterfaceSpec(nicSpec *instancegroup.NetworkInterfaceSpec) map[string]interface{} {
	nat := (nicSpec.PrimaryV4AddressSpec != nil && nicSpec.PrimaryV4AddressSpec.GetOneToOneNatSpec() != nil) ||
		(nicSpec.PrimaryV6AddressSpec != nil && nicSpec.PrimaryV6AddressSpec.GetOneToOneNatSpec() != nil)

	subnets := &schema.Set{F: schema.HashString}

	if nicSpec.SubnetIds != nil {
		for _, s := range nicSpec.SubnetIds {
			subnets.Add(s)
		}
	}

	networkInterface := map[string]interface{}{
		"network_id": nicSpec.NetworkId,
		"subnet_ids": subnets,
		"nat":        nat,
		"ipv4":       nicSpec.PrimaryV4AddressSpec != nil,
		"ipv6":       nicSpec.PrimaryV6AddressSpec != nil,
	}

	if nicSpec.GetSecurityGroupIds() != nil {
		networkInterface["security_group_ids"] = convertStringArrToInterface(nicSpec.SecurityGroupIds)
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
	}

	if sp := ig.GetScalePolicy().GetAutoScale(); sp != nil {
		res["auto_scale"], _ = flattenInstanceGroupAutoScale(sp)
	}

	if sp := ig.GetScalePolicy().GetTestAutoScale(); sp != nil {
		res["test_auto_scale"], _ = flattenInstanceGroupAutoScale(sp)
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupAutoScale(sp *instancegroup.ScalePolicy_AutoScale) ([]map[string]interface{}, error) {
	subres := map[string]interface{}{}
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
				"labels":      rule.GetLabels(),
			}
		}
	}

	return []map[string]interface{}{subres}, nil
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

	if _, ok := d.GetOk(prefix + ".image_id"); ok {
		if _, ok := d.GetOk(prefix + ".snapshot_id"); ok {
			return diskSpec, fmt.Errorf("Use one of  'image_id', 'snapshot_id', not both.")
		}
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

func expandPrimaryV4AddressSpec(config map[string]interface{}) (*compute.PrimaryAddressSpec, error) {
	if v, ok := config["ipv4"]; ok {
		if !v.(bool) {
			return nil, nil
		}

		natSpec, _ := expandOneToOneNatSpec(config)
		return &compute.PrimaryAddressSpec{
			Address:         config["ip_address"].(string),
			OneToOneNatSpec: natSpec,
		}, nil
	}
	return nil, nil
}

func expandPrimaryV6AddressSpec(config map[string]interface{}) (*compute.PrimaryAddressSpec, error) {
	if v, ok := config["ipv6"]; ok {
		if !v.(bool) {
			return nil, nil
		}

		return &compute.PrimaryAddressSpec{
			Address: config["ipv6_address"].(string),
		}, nil
	}
	return nil, nil
}

func expandSecurityGroupIds(v interface{}) []string {
	if v == nil {
		return nil
	}
	var m []string
	sgIdsSet := v.(*schema.Set)
	for _, val := range sgIdsSet.List() {
		m = append(m, val.(string))
	}
	return m
}

func expandOneToOneNatSpec(config map[string]interface{}) (*compute.OneToOneNatSpec, error) {
	if v, ok := config["nat"]; ok {
		if !v.(bool) {
			return nil, nil
		}

		if ipAddress, ok := config["nat_ip_address"].(string); ok && ipAddress != "" {
			return &compute.OneToOneNatSpec{
				Address: ipAddress,
			}, nil
		}
		return &compute.OneToOneNatSpec{
			IpVersion: compute.IpVersion_IPV4,
		}, nil
	}
	return nil, nil
}

func expandInstanceNetworkSettingsSpecs(d *schema.ResourceData) (*compute.NetworkSettings, error) {
	if v, ok := d.GetOk("network_acceleration_type"); ok {
		typeVal, ok := compute.NetworkSettings_Type_value[strings.ToUpper(v.(string))]
		if !ok {
			return nil, fmt.Errorf("value for 'network_acceleration_type' should be 'standard' or 'software_accelerated'', not '%s'", v)
		}
		return &compute.NetworkSettings{
			Type: compute.NetworkSettings_Type(typeVal),
		}, nil
	}
	return nil, nil
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

		if sgids, ok := data["security_group_ids"]; ok {
			nics[i].SecurityGroupIds = expandSecurityGroupIds(sgids)
		}

		ipV4Address := data["ip_address"].(string)
		ipV6Address := data["ipv6_address"].(string)

		// By default allocate any unassigned IPv4 address
		if ipV4Address == "" && ipV6Address == "" {
			nics[i].PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{}
		}

		if enableIPV4, ok := data["ipv4"].(bool); ok && enableIPV4 {
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
			natSpec := &compute.OneToOneNatSpec{
				IpVersion: compute.IpVersion_IPV4,
			}

			if natAddress, ok := data["nat_ip_address"].(string); ok && natAddress != "" {
				natSpec = &compute.OneToOneNatSpec{
					Address: natAddress,
				}
			}

			if nics[i].PrimaryV4AddressSpec == nil {
				nics[i].PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{
					OneToOneNatSpec: natSpec,
				}
			} else {
				nics[i].PrimaryV4AddressSpec.OneToOneNatSpec = natSpec
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

		if enableIPV4, ok := data["ipv4"].(bool); ok && enableIPV4 {
			nics[i].PrimaryV4AddressSpec = &instancegroup.PrimaryAddressSpec{}
		}

		if enableIPV6, ok := data["ipv6"].(bool); ok && enableIPV6 {
			nics[i].PrimaryV6AddressSpec = &instancegroup.PrimaryAddressSpec{}
		}

		if nat, ok := data["nat"].(bool); ok && nat {
			natSpec := &instancegroup.OneToOneNatSpec{
				IpVersion: instancegroup.IpVersion_IPV4,
			}
			if nics[i].PrimaryV4AddressSpec == nil {
				nics[i].PrimaryV4AddressSpec = &instancegroup.PrimaryAddressSpec{
					OneToOneNatSpec: natSpec,
				}
			} else {
				nics[i].PrimaryV4AddressSpec.OneToOneNatSpec = natSpec
			}
		}

		if sgids, ok := data["security_group_ids"]; ok {
			nics[i].SecurityGroupIds = expandSecurityGroupIds(sgids)
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
	val, ok := instancegroup.AttachedDiskSpec_Mode_value[mode]
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

func parseKmsDefaultAlgorithm(algo string) (kms.SymmetricAlgorithm, error) {
	val, ok := kms.SymmetricAlgorithm_value[algo]
	if !ok {
		return kms.SymmetricAlgorithm(0), fmt.Errorf("value for 'default_algorithm' should be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(kms.SymmetricAlgorithm_value)), algo)
	}
	return kms.SymmetricAlgorithm(val), nil
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

func flattenDhcpOptions(dhcpOptions *vpc.DhcpOptions) []interface{} {
	if dhcpOptions == nil {
		return nil
	}

	m := make(map[string]interface{})

	if dhcpOptions.DomainName != "" {
		m["domain_name"] = dhcpOptions.DomainName
	}

	if len(dhcpOptions.DomainNameServers) > 0 {
		m["domain_name_servers"] = dhcpOptions.DomainNameServers
	}

	if len(dhcpOptions.NtpServers) > 0 {
		m["ntp_servers"] = dhcpOptions.NtpServers
	}

	if len(m) > 0 {
		return []interface{}{m}
	}

	return nil
}

func expandDhcpOptions(d *schema.ResourceData) (*vpc.DhcpOptions, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("dhcp_options"); !ok {
		return nil, nil
	}

	optsList := v.([]interface{})

	if len(optsList) == 0 {
		return nil, nil
	}

	optsDescriptor, ok := optsList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", optsList[0])
	}

	var dhcpOptions vpc.DhcpOptions

	if v, ok := optsDescriptor["domain_name"].(string); ok {
		dhcpOptions.DomainName = v
	}

	if v, ok := optsDescriptor["domain_name_servers"].([]interface{}); ok {
		vs := make([]string, 0, len(v))
		for _, s := range v {
			vs = append(vs, s.(string))
		}
		dhcpOptions.DomainNameServers = vs
	}

	if v, ok := optsDescriptor["ntp_servers"].([]interface{}); ok {
		vs := make([]string, 0, len(v))
		for _, s := range v {
			vs = append(vs, s.(string))
		}
		dhcpOptions.NtpServers = vs
	}

	return &dhcpOptions, nil
}

// revive:disable:var-naming
func expandInstanceGroupInstanceTemplate(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.InstanceTemplate, error) {
	var platformId, description, serviceAccount, name, hostname string

	if v, ok := d.GetOk(prefix + ".platform_id"); ok {
		platformId = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".description"); ok {
		description = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".service_account_id"); ok {
		serviceAccount = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".name"); ok {
		name = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".hostname"); ok {
		hostname = v.(string)
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

	networkSettings, err := expandNetworkSettings(d.Get(prefix + ".network_settings.0.type"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding network settings while creating instance group: %s", err)
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
		NetworkSettings:       networkSettings,
		Name:                  name,
		Hostname:              hostname,
	}

	return template, nil
}

func expandInstanceGroupVariables(v interface{}) ([]*instancegroup.Variable, error) {
	variables := make([]*instancegroup.Variable, 0)
	if v == nil {
		return variables, nil
	}

	for key, val := range v.(map[string]interface{}) {
		variables = append(variables, &instancegroup.Variable{
			Key:   key,
			Value: val.(string),
		})
	}
	return variables, nil
}

func expandNetworkSettings(v interface{}) (*instancegroup.NetworkSettings, error) {
	ns := &instancegroup.NetworkSettings{}
	if v == nil || v.(string) == "" {
		return nil, nil
	}
	t, err := parseInstanceGroupNetworkSettingsType(v.(string))
	if err != nil {
		return nil, err
	}
	ns.Type = t
	return ns, nil
}

func parseInstanceGroupNetworkSettingsType(str string) (instancegroup.NetworkSettings_Type, error) {
	val, ok := instancegroup.NetworkSettings_Type_value[str]
	if !ok {
		return instancegroup.NetworkSettings_TYPE_UNSPECIFIED, fmt.Errorf("value for 'type' should be 'STANDARD' or 'SOFTWARE_ACCELERATED' or 'HARDWARE_ACCELERATED', not '%s'", str)
	}
	return instancegroup.NetworkSettings_Type(val), nil
}

func expandInstanceGroupScalePolicy(d *schema.ResourceData) (*instancegroup.ScalePolicy, error) {
	var policy = &instancegroup.ScalePolicy{}

	if _, ok := d.GetOk("scale_policy.0.fixed_scale"); ok {
		v := d.Get("scale_policy.0.fixed_scale.0.size").(int)
		policy.ScaleType = &instancegroup.ScalePolicy_FixedScale_{FixedScale: &instancegroup.ScalePolicy_FixedScale{Size: int64(v)}}
	}

	if _, ok := d.GetOk("scale_policy.0.auto_scale"); ok {
		autoScale, err := expandInstanceGroupAutoScale(d, "scale_policy.0.auto_scale.0")
		if err != nil {
			return nil, err
		}
		policy.ScaleType = &instancegroup.ScalePolicy_AutoScale_{AutoScale: autoScale}
		return policy, nil
	}

	if _, ok := d.GetOk("scale_policy.0.test_auto_scale"); ok {
		testAutoScale, err := expandInstanceGroupAutoScale(d, "scale_policy.0.test_auto_scale.0")
		if err != nil {
			return nil, err
		}
		policy.TestAutoScale = testAutoScale
		return policy, nil
	}

	if policy.ScaleType == nil {
		return nil, fmt.Errorf("Only fixed_scale and auto_scale policy are supported")
	}

	return policy, nil
}

func expandInstanceGroupAutoScale(d *schema.ResourceData, prefix string) (*instancegroup.ScalePolicy_AutoScale, error) {
	autoScale := &instancegroup.ScalePolicy_AutoScale{
		MinZoneSize: int64(d.Get(prefix + ".min_zone_size").(int)),
		MaxSize:     int64(d.Get(prefix + ".max_size").(int)),
		InitialSize: int64(d.Get(prefix + ".initial_size").(int)),
	}

	if v, ok := d.GetOk(prefix + ".measurement_duration"); ok {
		autoScale.MeasurementDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := d.GetOk(prefix + ".warmup_duration"); ok {
		autoScale.WarmupDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := d.GetOk(prefix + ".cpu_utilization_target"); ok {
		autoScale.CpuUtilizationRule = &instancegroup.ScalePolicy_CpuUtilizationRule{UtilizationTarget: v.(float64)}
	}

	if v, ok := d.GetOk(prefix + ".stabilization_duration"); ok {
		autoScale.StabilizationDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if customRulesCount := d.Get(prefix + ".custom_rule.#").(int); customRulesCount > 0 {
		rules := make([]*instancegroup.ScalePolicy_CustomRule, customRulesCount)
		for i := 0; i < customRulesCount; i++ {
			key := fmt.Sprintf(prefix+".custom_rule.%d", i)
			if rule, err := expandCustomRule(d, key); err == nil {
				rules[i] = rule
			} else {
				return nil, err
			}
		}
		autoScale.CustomRules = rules
	}

	return autoScale, nil
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

	labels, err := expandLabels(d.Get(prefix + ".labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating custom rule: %s", err)
	}

	return &instancegroup.ScalePolicy_CustomRule{
		RuleType:   instancegroup.ScalePolicy_CustomRule_RuleType(ruleType),
		MetricType: instancegroup.ScalePolicy_CustomRule_MetricType(metricType),
		MetricName: d.Get(prefix + ".metric_name").(string),
		Target:     d.Get(prefix + ".target").(float64),
		Labels:     labels,
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

func expandSecurityGroupRulesSpec(d *schema.ResourceData) ([]*vpc.SecurityGroupRuleSpec, error) {

	securityRules := make([]*vpc.SecurityGroupRuleSpec, 0)

	for _, dir := range []string{"egress", "ingress"} {
		if v, ok := d.GetOk(dir); ok {
			for _, rule := range v.(*schema.Set).List() {
				if r, err := securityRuleDescriptionToRuleSpec(dir, rule); err == nil {
					securityRules = append(securityRules, r)
				} else {
					return securityRules, err
				}
			}
		}
	}

	return securityRules, nil
}

func securityRuleDescriptionToRuleSpec(dir string, v interface{}) (*vpc.SecurityGroupRuleSpec, error) {
	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", v)
	}

	sr := new(vpc.SecurityGroupRuleSpec)

	directionId := vpc.SecurityGroupRule_Direction_value[strings.ToUpper(dir)]
	sr.SetDirection(vpc.SecurityGroupRule_Direction(directionId))

	if v, ok := res["description"].(string); ok {
		sr.Description = v
	}

	if protoName, protoNum, err := getProtocol(res["protocol"]); err == nil {
		if protoName != "" {
			sr.SetProtocolName(protoName)
		} else {
			sr.SetProtocolNumber(protoNum)
		}
	} else {
		return nil, err
	}

	if v, ok := res["labels"]; ok {
		labels, err := expandLabels(v)
		if err != nil {
			return sr, err
		}
		sr.Labels = labels
	}

	cidr := new(vpc.CidrBlocks)

	if v, ok := res["v4_cidr_blocks"]; ok {
		arr := v.([]interface{})
		cidr.V4CidrBlocks = make([]string, len(arr))
		for i, c := range arr {
			cidr.V4CidrBlocks[i] = c.(string)
		}
	}
	if v, ok := res["v6_cidr_blocks"]; ok {
		arr := v.([]interface{})
		cidr.V6CidrBlocks = make([]string, len(arr))
		for i, c := range arr {
			cidr.V6CidrBlocks[i] = c.(string)
		}
	}
	if cidr.V4CidrBlocks == nil && cidr.V6CidrBlocks == nil {
		return nil, fmt.Errorf("v4_cidr_blocks or v6_cidr_blocks must be defined")
	}
	sr.SetCidrBlocks(cidr)

	port := res["port"].(int)
	frPort := res["from_port"].(int)
	toPort := res["to_port"].(int)

	if port != -1 {
		if frPort != -1 || toPort != -1 {
			return nil, fmt.Errorf("port or from_port + to_port must be defined %v", res)
		}
		sr.SetPorts(&vpc.PortRange{FromPort: int64(port), ToPort: int64(port)})
	} else {
		if frPort == -1 || toPort == -1 {
			return nil, fmt.Errorf("port or from_port + to_port must be defined")
		}
		sr.SetPorts(&vpc.PortRange{FromPort: int64(frPort), ToPort: int64(toPort)})
	}

	return sr, nil
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

		changedAt, err := getTimestamp(instance.GetStatusChangedAt())
		if err != nil {
			return res, err
		}
		instDict["status_changed_at"] = changedAt

		networkInterfaces, _, _, err := flattenInstanceGroupManagedInstanceNetworkInterfaces(instance)
		if err != nil {
			return res, err
		}

		instDict["network_interface"] = networkInterfaces
		res[i] = instDict
	}

	return res, nil
}

func flattenMaintenanceWindow(mw *k8s.MaintenanceWindow) (*schema.Set, error) {
	maintenanceWindow := &schema.Set{F: dayOfWeekHash}
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *k8s.MaintenanceWindow_Anytime:
			// do nothing
		case *k8s.MaintenanceWindow_DailyMaintenanceWindow:
			dailyPolicy := map[string]interface{}{
				"start_time": formatTimeOfDay(p.DailyMaintenanceWindow.GetStartTime()),
				"duration":   formatDuration(p.DailyMaintenanceWindow.GetDuration()),
			}
			maintenanceWindow.Add(dailyPolicy)
		case *k8s.MaintenanceWindow_WeeklyMaintenanceWindow:
			for _, v := range p.WeeklyMaintenanceWindow.GetDaysOfWeek() {
				for _, d := range v.GetDays() {
					dailyPolicy := map[string]interface{}{
						"day":        strings.ToLower(d.String()),
						"start_time": formatTimeOfDay(v.GetStartTime()),
						"duration":   formatDuration(v.GetDuration()),
					}
					if maintenanceWindow.Contains(dailyPolicy) {
						return nil, fmt.Errorf("duplication for day of week found in maintenance window")
					}
					maintenanceWindow.Add(dailyPolicy)
				}
			}
		default:
			return nil, fmt.Errorf("unsupported Kubernetes master maintenance policy type")
		}
	}

	return maintenanceWindow, nil
}

func expandMaintenanceWindow(days []interface{}) (*k8s.MaintenanceWindow, error) {
	if len(days) == 0 {
		return nil, nil
	}

	windows := []*dayMaintenanceWindow{}
	parsedDays := map[dayofweek.DayOfWeek]struct{}{}
	dailyWindowSpecified := false

	for _, v := range days {
		window, err := expandDayMaintenanceWindow(v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		if window.day == dayofweek.DayOfWeek_DAY_OF_WEEK_UNSPECIFIED {
			dailyWindowSpecified = true
		}

		// duplicate day from config. can be either, any day, or specific day.
		if _, ok := parsedDays[window.day]; ok {
			return nil, fmt.Errorf("can not specify two time intervals for one day")
		}

		parsedDays[window.day] = struct{}{}
		windows = append(windows, window)
	}

	if dailyWindowSpecified {
		if len(windows) != 1 {
			return nil, fmt.Errorf("can not use daily and weekly maintenance window policies simultaneously")
		}

		return &k8s.MaintenanceWindow{
			Policy: &k8s.MaintenanceWindow_DailyMaintenanceWindow{
				DailyMaintenanceWindow: &k8s.DailyMaintenanceWindow{
					StartTime: windows[0].startTime,
					Duration:  windows[0].duration,
				},
			},
		}, nil
	}

	var dw []*k8s.DaysOfWeekMaintenanceWindow
	for _, w := range windows {
		dw = append(dw, &k8s.DaysOfWeekMaintenanceWindow{
			Days:      []dayofweek.DayOfWeek{w.day},
			StartTime: w.startTime,
			Duration:  w.duration,
		})
	}

	return &k8s.MaintenanceWindow{
		Policy: &k8s.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &k8s.WeeklyMaintenanceWindow{
				DaysOfWeek: dw,
			},
		},
	}, nil
}

func expandDayMaintenanceWindow(daySpec map[string]interface{}) (*dayMaintenanceWindow, error) {
	var (
		window dayMaintenanceWindow
		err    error
	)

	// special case. Terraform fills fields in Set resource, that are not present in original user config!
	if dayName, ok := daySpec["day"]; ok && dayName != "" {
		if window.day, err = parseDayOfWeek(dayName.(string)); err != nil {
			return nil, err
		}
	}

	if window.startTime, err = parseDayTime(daySpec["start_time"].(string)); err != nil {
		return nil, err
	}

	if window.duration, err = parseDuration(daySpec["duration"].(string)); err != nil {
		return nil, err
	}

	return &window, nil
}

type dayMaintenanceWindow struct {
	day       dayofweek.DayOfWeek
	startTime *timeofday.TimeOfDay
	duration  *duration.Duration
}

func expandDataprocCreateClusterConfigSpec(d *schema.ResourceData) *dataproc.CreateClusterConfigSpec {
	return &dataproc.CreateClusterConfigSpec{
		VersionId:       d.Get("cluster_config.0.version_id").(string),
		Hadoop:          expandDataprocHadoopConfig(d),
		SubclustersSpec: expandDataprocSubclustersSpec(d),
	}
}

func expandDataprocHadoopConfig(d *schema.ResourceData) *dataproc.HadoopConfig {
	return &dataproc.HadoopConfig{
		Services:      expandDataprocServices(d),
		Properties:    expandDataprocProperties(d),
		SshPublicKeys: expandDataprocSSHPublicKeys(d),
	}
}

func expandDataprocServices(d *schema.ResourceData) []dataproc.HadoopConfig_Service {
	set := d.Get("cluster_config.0.hadoop.0.services").(*schema.Set)
	serviceNames := convertStringSet(set)
	sort.Strings(serviceNames)
	services := make([]dataproc.HadoopConfig_Service, len(serviceNames))

	for i, serviceName := range serviceNames {
		// service name is checked by validation
		serviceID := dataproc.HadoopConfig_Service_value[serviceName]
		services[i] = dataproc.HadoopConfig_Service(serviceID)
	}

	return services
}

func expandDataprocProperties(d *schema.ResourceData) map[string]string {
	v := d.Get("cluster_config.0.hadoop.0.properties").(map[string]interface{})
	return convertStringMap(v)
}

func expandDataprocSSHPublicKeys(d *schema.ResourceData) []string {
	v := d.Get("cluster_config.0.hadoop.0.ssh_public_keys").(*schema.Set)
	return convertStringSet(v)
}

func expandDataprocSubclustersSpec(d *schema.ResourceData) []*dataproc.CreateSubclusterConfigSpec {
	rootKey := "cluster_config.0.subcluster_spec"
	list := d.Get(rootKey).([]interface{})
	subclusters := make([]*dataproc.CreateSubclusterConfigSpec, len(list))
	for index, element := range list {
		subclusters[index] = expandDataprocSubclusterSpec(element)
	}

	return subclusters
}

func expandDataprocSubclusterSpec(element interface{}) *dataproc.CreateSubclusterConfigSpec {
	subclusterSpec := element.(map[string]interface{})
	roleName := subclusterSpec["role"].(string)
	roleID := dataproc.Role_value[roleName]
	resourcesSpec := subclusterSpec["resources"].([]interface{})[0]

	return &dataproc.CreateSubclusterConfigSpec{
		Role:       dataproc.Role(roleID),
		Name:       subclusterSpec["name"].(string),
		SubnetId:   subclusterSpec["subnet_id"].(string),
		HostsCount: int64(subclusterSpec["hosts_count"].(int)),
		Resources:  expandDataprocResources(resourcesSpec),
	}
}

func expandDataprocResources(r interface{}) *dataproc.Resources {
	resources := &dataproc.Resources{}
	resourcesMap := r.(map[string]interface{})

	if v, ok := resourcesMap["resource_preset_id"]; ok {
		resources.ResourcePresetId = v.(string)
	}
	if v, ok := resourcesMap["disk_size"]; ok {
		resources.DiskSize = toBytes(v.(int))
	}
	if v, ok := resourcesMap["disk_type_id"]; ok {
		resources.DiskTypeId = v.(string)
	}
	return resources
}

func flattenDataprocClusterConfig(cluster *dataproc.Cluster, subclusters []*dataproc.Subcluster) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"version_id":      cluster.Config.VersionId,
			"hadoop":          flattenDataprocHadoopConfig(cluster.Config.Hadoop),
			"subcluster_spec": flattenDataprocSubclusters(subclusters),
		},
	}
}

func flattenDataprocHadoopConfig(config *dataproc.HadoopConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"services":        flattenDataprocServices(config.Services),
			"properties":      config.Properties,
			"ssh_public_keys": config.SshPublicKeys,
		},
	}
}

func flattenDataprocServices(services []dataproc.HadoopConfig_Service) []string {
	serviceNames := make([]string, len(services))
	for idx, service := range services {
		serviceNames[idx] = service.String()
	}
	return serviceNames
}

func flattenDataprocSubclusters(subclusters []*dataproc.Subcluster) []interface{} {
	result := make([]interface{}, len(subclusters))
	for idx, subcluster := range subclusters {
		result[idx] = flattenDataprocSubcluster(subcluster)
	}
	return result
}

func flattenDataprocSubcluster(subcluster *dataproc.Subcluster) map[string]interface{} {
	return map[string]interface{}{
		"id":          subcluster.Id,
		"name":        subcluster.Name,
		"role":        subcluster.Role.String(),
		"resources":   flattenDataprocResources(subcluster.Resources),
		"subnet_id":   subcluster.SubnetId,
		"hosts_count": subcluster.HostsCount,
	}
}

func flattenDataprocResources(r *dataproc.Resources) []map[string]interface{} {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return []map[string]interface{}{res}
}

func flattenSecurityGroupRulesSpec(sg []*vpc.SecurityGroupRule) (*schema.Set, *schema.Set) {
	ingress := schema.NewSet(resourceYandexVPCSecurityGroupRuleHash, nil)
	egress := schema.NewSet(resourceYandexVPCSecurityGroupRuleHash, nil)

	for _, g := range sg {
		r := make(map[string]interface{})
		r["description"] = g.GetDescription()
		r["labels"] = g.GetLabels()

		if g.GetProtocolNumber() == 0 {
			r["protocol"] = "ANY"
		} else {
			found := false

			for _, s := range validProtocols {
				if g.GetProtocolName() == s {
					r["protocol"] = s
					found = true
					break
				}
			}

			if !found {
				r["protocol"] = fmt.Sprintf("%d", g.GetProtocolNumber())
			}
		}

		if g.GetPorts() != nil {
			if g.GetPorts().FromPort == g.GetPorts().ToPort {
				r["port"] = g.GetPorts().FromPort
				r["from_port"] = -1
				r["to_port"] = -1
			} else {
				r["port"] = -1
				r["from_port"] = g.GetPorts().FromPort
				r["to_port"] = g.GetPorts().ToPort
			}
		}

		if g.GetCidrBlocks() != nil && g.GetCidrBlocks().V4CidrBlocks != nil {
			r["v4_cidr_blocks"] = convertStringArrToInterface(g.GetCidrBlocks().V4CidrBlocks)
		}

		if g.GetCidrBlocks() != nil && g.GetCidrBlocks().V6CidrBlocks != nil {
			r["v6_cidr_blocks"] = convertStringArrToInterface(g.GetCidrBlocks().V6CidrBlocks)
		}

		r["id"] = g.Id

		switch g.GetDirection().String() {
		case "INGRESS":
			ingress.Add(r)
		case "EGRESS":
			egress.Add(r)
		}

	}
	return ingress, egress
}
