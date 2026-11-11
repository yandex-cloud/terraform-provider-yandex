package pluginsdk

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc"
	proto "google.golang.org/protobuf/proto"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

type ReducedDiskServiceClient interface {
	Get(ctx context.Context, in *compute.GetDiskRequest, opts ...grpc.CallOption) (*compute.Disk, error)
}

func ExpandProductIds(v interface{}) ([]string, error) {
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

func expandHardwareGeneration(d *schema.ResourceData) (*compute.HardwareGeneration, error) {
	val := new(compute.HardwareGeneration)

	if _, ok := d.GetOk("hardware_generation.0.legacy_features"); ok {
		legacyFeatures, err := expandDiskHardwareGenerationLegacyFeatures(d)
		if err != nil {
			return nil, err
		}

		val.SetLegacyFeatures(legacyFeatures)
	}

	if _, ok := d.GetOk("hardware_generation.0.generation2_features"); ok {
		generation2Features, err := expandDiskHardwareGenerationGeneration2Features(d)
		if err != nil {
			return nil, err
		}

		val.SetGeneration2Features(generation2Features)
	}

	empty := new(compute.HardwareGeneration)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandDiskHardwareGenerationLegacyFeatures(d *schema.ResourceData) (*compute.LegacyHardwareFeatures, error) {
	val := new(compute.LegacyHardwareFeatures)

	if v, ok := d.GetOk("hardware_generation.0.legacy_features.0.pci_topology"); ok {
		pciTopology, err := parseComputePCITopology(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPciTopology(pciTopology)
	}

	return val, nil
}

func expandDiskHardwareGenerationGeneration2Features(_ *schema.ResourceData) (*compute.Generation2HardwareFeatures, error) {
	val := new(compute.Generation2HardwareFeatures)
	return val, nil
}

func flattenComputeHardwareGeneration(v *compute.HardwareGeneration) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	generation2Features, err := flattenComputeHardwareGenerationgeneration2Features(v.GetGeneration2Features())
	if err != nil {
		return nil, err
	}
	m["generation2_features"] = generation2Features
	legacyFeatures, err := flattenComputeHardwareGenerationlegacyFeatures(v.GetLegacyFeatures())
	if err != nil {
		return nil, err
	}
	m["legacy_features"] = legacyFeatures

	return []map[string]interface{}{m}, nil
}

func flattenComputeHardwareGenerationgeneration2Features(v *compute.Generation2HardwareFeatures) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})
	return []map[string]interface{}{m}, nil
}

func flattenComputeHardwareGenerationlegacyFeatures(v *compute.LegacyHardwareFeatures) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["pci_topology"] = v.PciTopology.String()

	return []map[string]interface{}{m}, nil
}

func FlattenInstanceResources(instance *compute.Instance) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"cores":         int(instance.Resources.Cores),
		"core_fraction": int(instance.Resources.CoreFraction),
		"memory":        sdkcommon.ToGigabytesInFloat(instance.Resources.Memory),
		"gpus":          int(instance.Resources.Gpus),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func FlattenInstanceBootDisk(ctx context.Context, instance *compute.Instance, diskServiceClient ReducedDiskServiceClient) ([]map[string]interface{}, error) {
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

	var kmsKey string
	if disk.KmsKey != nil {
		kmsKey = disk.KmsKey.KeyId
	}

	bootDisk["initialize_params"] = []map[string]interface{}{{
		"name":        disk.Name,
		"description": disk.Description,
		"size":        sdkcommon.ToGigabytes(disk.Size),
		"block_size":  int(disk.BlockSize),
		"type":        disk.TypeId,
		"image_id":    disk.GetSourceImageId(),
		"snapshot_id": disk.GetSourceSnapshotId(),
		"kms_key_id":  kmsKey,
	}}

	return []map[string]interface{}{bootDisk}, nil
}

func flattenInstanceSecondaryDisks(instance *compute.Instance) ([]map[string]interface{}, error) {
	var secondaryDisks []map[string]interface{}

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

func hashInstanceSecondaryDisks(v interface{}) int {
	var buf bytes.Buffer

	disk := v.(map[string]interface{})
	if v, ok := disk["disk_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := disk["mode"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := disk["auto_delete"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", v.(bool)))
	}

	return hashcode.String(buf.String())

}

func FlattenInstanceNetworkInterfaces(instance *compute.Instance) ([]map[string]interface{}, string, string, error) {
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
			nics[i]["security_group_ids"] = sdkcommon.ConvertStringArrToInterface(iface.SecurityGroupIds)
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

		if sp := iface.GetPrimaryV4Address().GetDnsRecords(); sp != nil {
			nics[i]["dns_record"] = flattenComputeInstanceDnsRecords(sp)
		}

		if sp := iface.GetPrimaryV6Address().GetDnsRecords(); sp != nil {
			nics[i]["ipv6_dns_record"] = flattenComputeInstanceDnsRecords(sp)
		}

		if sp := iface.GetPrimaryV4Address().GetOneToOneNat().GetDnsRecords(); sp != nil {
			nics[i]["nat_dns_record"] = flattenComputeInstanceDnsRecords(sp)
		}
	}

	return nics, externalIP, internalIP, nil
}

func flattenComputeInstanceDnsRecords(specs []*compute.DnsRecord) []map[string]interface{} {
	res := make([]map[string]interface{}, len(specs))

	for i, spec := range specs {
		res[i] = map[string]interface{}{
			"fqdn":        spec.Fqdn,
			"dns_zone_id": spec.DnsZoneId,
			"ttl":         int(spec.Ttl),
			"ptr":         spec.Ptr,
		}
	}

	return res
}

func flattenInstanceFilesystems(instance *compute.Instance) []map[string]interface{} {
	filesystems := make([]map[string]interface{}, len(instance.Filesystems))

	for i, spec := range instance.Filesystems {
		filesystems[i] = map[string]interface{}{
			"filesystem_id": spec.FilesystemId,
			"device_name":   spec.DeviceName,
			"mode":          spec.GetMode().String(),
		}
	}
	return filesystems
}

func hashFilesystem(v interface{}) int {
	var buf bytes.Buffer

	fs := v.(map[string]interface{})
	if v, ok := fs["filesystem_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := fs["mode"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())

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
		rs.Memory = sdkcommon.ToBytesFromFloat(v.(float64))
	}

	return rs, nil
}

func expandInstanceBootDiskSpec(d *schema.ResourceData, config *providerconfig.Config) (*compute.AttachedDiskSpec, error) {
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

func expandBootDiskSpec(d *schema.ResourceData, config *providerconfig.Config) (*compute.AttachedDiskSpec_DiskSpec, error) {
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
		diskSpec.Size = sdkcommon.ToBytes(v.(int))
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.block_size"); ok {
		diskSpec.BlockSize = int64(v.(int))
	}

	if v, ok := d.GetOk("boot_disk.0.initialize_params.0.kms_key_id"); ok {
		diskSpec.KmsKeyId = v.(string)
	}

	if diskSpec.Size == 0 {
		diskSpec.Size = minStorageSizeBytes
	}

	return diskSpec, nil
}

func expandInstanceSecondaryDiskSpecs(d *schema.ResourceData) ([]*compute.AttachedDiskSpec, error) {
	ads := make([]*compute.AttachedDiskSpec, 0)

	secondaryDisks := d.Get("secondary_disk").(*schema.Set).List()
	for _, disk := range secondaryDisks {
		diskConfig := disk.(map[string]interface{})

		diskSpec, err := expandSecondaryDiskSpec(diskConfig)
		if err != nil {
			return nil, err
		}
		ads = append(ads, diskSpec)
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

func ExpandPrimaryV4AddressSpec(config map[string]interface{}) (*compute.PrimaryAddressSpec, error) {
	if v, ok := config["ipv4"]; ok {
		if !v.(bool) {
			return nil, nil
		}

		natSpec, _ := expandOneToOneNatSpec(config)

		var dnsSpecs []*compute.DnsRecordSpec

		if v1, ok := config["dns_record"]; ok {
			dnsSpecs = expandComputeInstanceDnsRecords(v1.([]interface{}))
		}

		return &compute.PrimaryAddressSpec{
			Address:         config["ip_address"].(string),
			OneToOneNatSpec: natSpec,
			DnsRecordSpecs:  dnsSpecs,
		}, nil
	}
	return nil, nil
}

func expandPrimaryV6AddressSpec(config map[string]interface{}) (*compute.PrimaryAddressSpec, error) {
	if v, ok := config["ipv6"]; ok {
		if !v.(bool) {
			return nil, nil
		}

		var dnsSpecs []*compute.DnsRecordSpec

		if v1, ok := config["ipv6_dns_record"]; ok {
			dnsSpecs = expandComputeInstanceDnsRecords(v1.([]interface{}))
		}

		return &compute.PrimaryAddressSpec{
			Address:        config["ipv6_address"].(string),
			DnsRecordSpecs: dnsSpecs,
		}, nil
	}
	return nil, nil
}

func expandNetworkInterfaceSpec(data map[string]interface{}) (*compute.NetworkInterfaceSpec, error) {
	subnetID := data["subnet_id"].(string)
	if subnetID == "" {
		return nil, fmt.Errorf("does not have a 'subnet_id' attribute defined")
	}

	iface := &compute.NetworkInterfaceSpec{
		SubnetId: subnetID,
	}

	if sgids, ok := data["security_group_ids"]; ok {
		iface.SecurityGroupIds = sdkcommon.ExpandSecurityGroupIds(sgids)
	}

	ipV4Address := data["ip_address"].(string)
	ipV6Address := data["ipv6_address"].(string)
	enableIPV4, ipv4Ok := data["ipv4"].(bool)

	// By default allocate any unassigned IPv4 address
	if ipV4Address == "" && ipV6Address == "" && enableIPV4 {
		iface.PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{}
	}

	if ipv4Ok && enableIPV4 {
		iface.PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{}
	}

	if enableIPV6, ok := data["ipv6"].(bool); ok && enableIPV6 {
		iface.PrimaryV6AddressSpec = &compute.PrimaryAddressSpec{}
	}

	if ipV4Address != "" {
		iface.PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{
			Address: ipV4Address,
		}
	}

	if ipV6Address != "" {
		iface.PrimaryV6AddressSpec = &compute.PrimaryAddressSpec{
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

		if iface.PrimaryV4AddressSpec == nil {
			iface.PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{
				OneToOneNatSpec: natSpec,
			}
		} else {
			iface.PrimaryV4AddressSpec.OneToOneNatSpec = natSpec
		}
	}

	if rec, ok := data["dns_record"]; ok {
		if iface.PrimaryV4AddressSpec != nil {
			iface.PrimaryV4AddressSpec.DnsRecordSpecs = expandComputeInstanceDnsRecords(rec.([]interface{}))
		}
	}

	if rec, ok := data["ipv6_dns_record"]; ok {
		if iface.PrimaryV6AddressSpec != nil {
			iface.PrimaryV6AddressSpec.DnsRecordSpecs = expandComputeInstanceDnsRecords(rec.([]interface{}))
		}
	}

	if rec, ok := data["nat_dns_record"]; ok {
		if iface.PrimaryV4AddressSpec != nil && iface.PrimaryV4AddressSpec.OneToOneNatSpec != nil {
			iface.PrimaryV4AddressSpec.OneToOneNatSpec.DnsRecordSpecs = expandComputeInstanceDnsRecords(rec.([]interface{}))
		}
	}
	return iface, nil
}

func expandOneToOneNatSpec(config map[string]interface{}) (*compute.OneToOneNatSpec, error) {
	if v, ok := config["nat"]; ok {
		if !v.(bool) {
			return nil, nil
		}
		var dnsSpecs []*compute.DnsRecordSpec

		if v1, ok := config["nat_dns_record"]; ok {
			dnsSpecs = expandComputeInstanceDnsRecords(v1.([]interface{}))
		}

		if ipAddress, ok := config["nat_ip_address"].(string); ok && ipAddress != "" {
			return &compute.OneToOneNatSpec{
				IpVersion:      compute.IpVersion_IPV4,
				Address:        ipAddress,
				DnsRecordSpecs: dnsSpecs,
			}, nil
		}
		return &compute.OneToOneNatSpec{
			IpVersion:      compute.IpVersion_IPV4,
			DnsRecordSpecs: dnsSpecs,
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

		iface, err := expandNetworkInterfaceSpec(data)
		if err != nil {
			return nil, fmt.Errorf("Failed to process NIC number #%d:%v", i, err)
		}
		nics[i] = iface
	}

	return nics, nil
}

func expandComputeInstanceDnsRecords(data []interface{}) []*compute.DnsRecordSpec {
	recs := make([]*compute.DnsRecordSpec, len(data))

	for i, raw := range data {
		d := raw.(map[string]interface{})
		r := &compute.DnsRecordSpec{Fqdn: d["fqdn"].(string)}
		if s, ok := d["dns_zone_id"]; ok {
			r.DnsZoneId = s.(string)
		}
		if s, ok := d["ttl"]; ok {
			r.Ttl = int64(s.(int))
		}
		if s, ok := d["ptr"]; ok {
			r.Ptr = s.(bool)
		}
		recs[i] = r
	}

	return recs
}

func parseDiskMode(mode string) (compute.AttachedDiskSpec_Mode, error) {
	val, ok := compute.AttachedDiskSpec_Mode_value[mode]
	if !ok {
		return compute.AttachedDiskSpec_MODE_UNSPECIFIED, fmt.Errorf("value for 'mode' should be 'READ_WRITE' or 'READ_ONLY', not '%s'", mode)
	}
	return compute.AttachedDiskSpec_Mode(val), nil
}

func parseFilesystemMode(mode string) (compute.AttachedFilesystemSpec_Mode, error) {
	val, ok := compute.AttachedFilesystemSpec_Mode_value[mode]
	if !ok {
		return compute.AttachedFilesystemSpec_MODE_UNSPECIFIED, fmt.Errorf("value for 'mode' should be 'READ_WRITE' or 'READ_ONLY', not '%s'", mode)
	}
	return compute.AttachedFilesystemSpec_Mode(val), nil
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

func expandInstancePlacementPolicy(d *schema.ResourceData) (*compute.PlacementPolicy, error) {
	sp := d.Get("placement_policy").([]interface{})
	var placementPolicy *compute.PlacementPolicy
	if len(sp) == 0 {
		return placementPolicy, nil
	}

	ruleSpecs := d.Get("placement_policy.0.host_affinity_rules").([]interface{})
	placementPolicy = &compute.PlacementPolicy{
		PlacementGroupId:  d.Get("placement_policy.0.placement_group_id").(string),
		HostAffinityRules: ExpandHostAffinityRulesSpec(ruleSpecs),
	}
	p, ok := d.GetOk("placement_policy.0.placement_group_partition")
	if ok {
		partitionNumber := p.(int)
		placementPolicy.PlacementGroupPartition = int64(partitionNumber)
	}
	return placementPolicy, nil
}

func expandInstanceMetadataOptions(d *schema.ResourceData) *compute.MetadataOptions {
	metadataOptions := compute.MetadataOptions{}
	if v, ok := d.GetOk("metadata_options.0.gce_http_endpoint"); ok {
		metadataOptions.GceHttpEndpoint = compute.MetadataOption(v.(int))
	}
	if v, ok := d.GetOk("metadata_options.0.aws_v1_http_endpoint"); ok {
		metadataOptions.AwsV1HttpEndpoint = compute.MetadataOption(v.(int))
	}
	if v, ok := d.GetOk("metadata_options.0.gce_http_token"); ok {
		metadataOptions.GceHttpToken = compute.MetadataOption(v.(int))
	}
	if v, ok := d.GetOk("metadata_options.0.aws_v1_http_token"); ok {
		metadataOptions.AwsV1HttpToken = compute.MetadataOption(v.(int))
	}
	if v, ok := d.GetOk("metadata_options.0.aws_v2_http_endpoint"); ok {
		metadataOptions.AwsV2HttpEndpoint = compute.MetadataOption(v.(int))
	}
	if v, ok := d.GetOk("metadata_options.0.aws_v2_http_token"); ok {
		metadataOptions.AwsV2HttpToken = compute.MetadataOption(v.(int))
	}

	return &metadataOptions
}

func ExpandHostAffinityRulesSpec(ruleSpecs []interface{}) []*compute.PlacementPolicy_HostAffinityRule {
	rulesCount := len(ruleSpecs)
	hostAffinityRules := make([]*compute.PlacementPolicy_HostAffinityRule, rulesCount)
	for i := 0; i < rulesCount; i++ {
		ruleSpec := ruleSpecs[i].(map[string]interface{})
		operator := compute.PlacementPolicy_HostAffinityRule_Operator_value[ruleSpec["op"].(string)]

		var values []string
		for _, value := range ruleSpec["values"].([]interface{}) {
			values = append(values, value.(string))
		}
		hostAffinityRules[i] = &compute.PlacementPolicy_HostAffinityRule{
			Key:    ruleSpec["key"].(string),
			Op:     compute.PlacementPolicy_HostAffinityRule_Operator(operator),
			Values: values,
		}
	}
	return hostAffinityRules
}

func ExpandLocalDiskSpecs(disks interface{}) []*compute.AttachedLocalDiskSpec {
	if disks == nil {
		return nil
	}

	var localDiskSpecs []*compute.AttachedLocalDiskSpec
	diskSpecs := disks.([]interface{})
	if len(diskSpecs) != 0 {
		localDiskSpecs = make([]*compute.AttachedLocalDiskSpec, 0, len(diskSpecs))
		for _, spec := range diskSpecs {
			diskSpec := spec.(map[string]interface{})
			diskSize := int64(diskSpec["size_bytes"].(int))
			localDisk := &compute.AttachedLocalDiskSpec{
				Size: diskSize,
				Type: &compute.AttachedLocalDiskSpec_PhysicalLocalDisk{
					PhysicalLocalDisk: &compute.PhysicalLocalDiskSpec{},
				},
			}

			if kmsKeyID, ok := diskSpec["kms_key_id"]; ok {
				localDisk.GetPhysicalLocalDisk().KmsKeyId = kmsKeyID.(string)
			}

			localDiskSpecs = append(localDiskSpecs, localDisk)
		}
	}
	return localDiskSpecs
}

func expandFilesystemSpec(fsConfig map[string]interface{}) (*compute.AttachedFilesystemSpec, error) {
	fs := compute.AttachedFilesystemSpec{}

	if v, ok := fsConfig["filesystem_id"]; ok {
		fs.FilesystemId = v.(string)
	}

	if v, ok := fsConfig["mode"]; ok {
		mode, err := parseFilesystemMode(v.(string))
		if err != nil {
			return nil, err
		}
		fs.Mode = mode
	}

	if v, ok := fsConfig["device_name"]; ok {
		fs.DeviceName = v.(string)
	}

	return &fs, nil
}

func expandInstanceFilesystemSpecs(d *schema.ResourceData) ([]*compute.AttachedFilesystemSpec, error) {
	var fsSpecs []*compute.AttachedFilesystemSpec

	specs := d.Get("filesystem").(*schema.Set)
	for _, spec := range specs.List() {
		fsConfig := spec.(map[string]interface{})
		log.Printf("# FS config: %+v", fsConfig)

		fs, err := expandFilesystemSpec(fsConfig)
		if err != nil {
			return nil, err
		}
		log.Printf("# FS spec: %+v", fs)
		fsSpecs = append(fsSpecs, fs)
	}
	return fsSpecs, nil
}

func expandInstanceGpuSettingsSpec(d *schema.ResourceData) (*compute.GpuSettings, error) {
	if v, ok := d.GetOk("gpu_cluster_id"); ok {
		return &compute.GpuSettings{
			GpuClusterId: v.(string),
		}, nil
	}
	return nil, nil
}

func flattenInstanceSchedulingPolicy(instance *compute.Instance) ([]map[string]interface{}, error) {
	schedulingPolicy := make([]map[string]interface{}, 0, 1)
	schedulingMap := map[string]interface{}{
		"preemptible": instance.SchedulingPolicy.Preemptible,
	}
	schedulingPolicy = append(schedulingPolicy, schedulingMap)
	return schedulingPolicy, nil
}

func flattenInstancePlacementPolicy(instance *compute.Instance) ([]map[string]interface{}, error) {
	placementPolicy := make([]map[string]interface{}, 0, 1)
	var affinityRules []interface{}
	for _, rule := range instance.PlacementPolicy.HostAffinityRules {
		affinityRules = append(affinityRules, map[string]interface{}{
			"key":    rule.Key,
			"op":     rule.Op.String(),
			"values": rule.Values,
		})
	}
	placementMap := map[string]interface{}{
		"placement_group_id":        instance.PlacementPolicy.PlacementGroupId,
		"placement_group_partition": instance.PlacementPolicy.PlacementGroupPartition,
		"host_affinity_rules":       affinityRules,
	}
	placementPolicy = append(placementPolicy, placementMap)
	return placementPolicy, nil
}

func flattenInstanceMetadataOptions(instance *compute.Instance) []map[string]interface{} {
	metadataOptions := map[string]interface{}{
		"gce_http_endpoint":    int(instance.MetadataOptions.GceHttpEndpoint),
		"aws_v1_http_endpoint": int(instance.MetadataOptions.AwsV1HttpEndpoint),
		"gce_http_token":       int(instance.MetadataOptions.GceHttpToken),
		"aws_v1_http_token":    int(instance.MetadataOptions.AwsV1HttpToken),
		"aws_v2_http_endpoint": int(instance.MetadataOptions.AwsV2HttpEndpoint),
		"aws_v2_http_token":    int(instance.MetadataOptions.AwsV2HttpToken),
	}

	return []map[string]interface{}{metadataOptions}
}

func FlattenLocalDisks(instance *compute.Instance) []interface{} {
	if len(instance.LocalDisks) == 0 {
		return nil
	}
	result := make([]interface{}, len(instance.LocalDisks))
	for i, disk := range instance.LocalDisks {
		spec := map[string]interface{}{
			"size_bytes":  int(disk.Size),
			"device_name": disk.DeviceName,
		}

		if disk.GetPhysicalLocalDisk() != nil && disk.GetPhysicalLocalDisk().KmsKey != nil {
			spec["kms_key_id"] = disk.GetPhysicalLocalDisk().KmsKey.KeyId
		}

		result[i] = spec
	}
	return result
}

func FlattenSnapshotScheduleSchedulePolicy(policy *compute.SchedulePolicy) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"expression": policy.GetExpression(),
		"start_at":   sdkcommon.GetTimestamp(policy.GetStartAt()),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func expandSnapshotScheduleSchedulePolicy(d *schema.ResourceData) (*compute.SchedulePolicy, error) {
	val := new(compute.SchedulePolicy)

	if v, ok := d.GetOk("schedule_policy.0.start_at"); ok {
		startAt, err := sdkcommon.ParseTimestamp(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetStartAt(startAt)
	}

	if v, ok := d.GetOk("schedule_policy.0.expression"); ok {
		val.SetExpression(v.(string))
	}

	empty := new(compute.SchedulePolicy)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func FlattenSnapshotScheduleSnapshotSpec(spec *compute.SnapshotSpec) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"description": spec.GetDescription(),
		"labels":      spec.GetLabels(),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func expandSnapshotScheduleSnapshotSpec(d *schema.ResourceData) (*compute.SnapshotSpec, error) {
	val := new(compute.SnapshotSpec)

	if v, ok := d.GetOk("snapshot_spec.0.description"); ok {
		val.SetDescription(v.(string))
	}

	if v, ok := d.GetOk("snapshot_spec.0.labels"); ok {
		labels, err := sdkcommon.ExpandLabels(v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLabels(labels)
	}

	empty := new(compute.SnapshotSpec)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}
