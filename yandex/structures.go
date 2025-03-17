package yandex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	kmsasymmetricencryption "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	kmsasymmetricsignature "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	ltagent "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1/agent"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/grpc"
	proto "google.golang.org/protobuf/proto"
)

func IterateKeys(d *schema.ResourceData, key string) []string {
	size := d.Get(key + ".#").(int)
	var keys []string
	for i := 0; i < size; i++ {
		currentKey := fmt.Sprintf(key+".%d.", i)
		keys = append(keys, currentKey)
	}
	return keys
}

type ReducedDiskServiceClient interface {
	Get(ctx context.Context, in *compute.GetDiskRequest, opts ...grpc.CallOption) (*compute.Disk, error)
}

func expandStringSet(v interface{}) []string {
	if v == nil {
		return nil
	}
	var result []string
	set := v.(*schema.Set)
	for _, val := range set.List() {
		result = append(result, val.(string))
	}
	return result
}

func expandStringSlice(v []interface{}) []string {
	if v == nil {
		return nil
	}
	s := make([]string, len(v))
	for i, val := range v {
		s[i] = val.(string)
	}
	return s
}

func expandStringStringMap(v map[string]interface{}) map[string]string {
	m := make(map[string]string)
	if v == nil {
		return m
	}
	for k, val := range v {
		m[k] = val.(string)
	}
	return m
}

func expandInt64Slice(v []interface{}) []int64 {
	if v == nil {
		return nil
	}
	s := make([]int64, len(v))
	for i, val := range v {
		s[i] = int64(val.(int))
	}
	return s
}

func expandLabels(v interface{}) (map[string]string, error) {
	m := make(map[string]string)
	if v == nil {
		return m, nil
	}
	for k, val := range v.(map[string]interface{}) {
		m[k] = strings.ToLower(val.(string))
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

func flattenInstanceResources(instance *compute.Instance) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"cores":         int(instance.Resources.Cores),
		"core_fraction": int(instance.Resources.CoreFraction),
		"memory":        toGigabytesInFloat(instance.Resources.Memory),
		"gpus":          int(instance.Resources.Gpus),
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

	var kmsKey string
	if disk.KmsKey != nil {
		kmsKey = disk.KmsKey.KeyId
	}

	bootDisk["initialize_params"] = []map[string]interface{}{{
		"name":        disk.Name,
		"description": disk.Description,
		"size":        toGigabytes(disk.Size),
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

func flattenK8SNodeGroupDNSRecords(specs []*k8s.DnsRecordSpec) []map[string]interface{} {
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

func expandPrimaryV4AddressSpec(config map[string]interface{}) (*compute.PrimaryAddressSpec, error) {
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
		iface.SecurityGroupIds = expandSecurityGroupIds(sgids)
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

func expandHostGroupIds(v interface{}) []string {
	if v == nil {
		return nil
	}
	var m []string
	hgIdsSet := v.(*schema.Set)
	for _, val := range hgIdsSet.List() {
		m = append(m, val.(string))
	}
	return m
}

func expandSubnetIds(v interface{}) []string {
	if v == nil {
		return nil
	}
	var m []string
	subnetIdsSet := v.(*schema.Set)
	for _, val := range subnetIdsSet.List() {
		m = append(m, val.(string))
	}
	return m
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

func expandK8SNodeGroupDNSRecords(data []interface{}) []*k8s.DnsRecordSpec {
	recs := make([]*k8s.DnsRecordSpec, len(data))

	for i, raw := range data {
		d := raw.(map[string]interface{})
		r := &k8s.DnsRecordSpec{Fqdn: d["fqdn"].(string)}
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

func parseKmsAsymmetricEncryptionAlgorithm(algo string) (kmsasymmetricencryption.AsymmetricEncryptionAlgorithm, error) {
	val, ok := kmsasymmetricencryption.AsymmetricEncryptionAlgorithm_value[algo]
	if !ok {
		return kmsasymmetricencryption.AsymmetricEncryptionAlgorithm(0), fmt.Errorf("value for 'encryption_algorithm' should be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(kmsasymmetricencryption.AsymmetricEncryptionAlgorithm_value)), algo)
	}
	return kmsasymmetricencryption.AsymmetricEncryptionAlgorithm(val), nil
}

func parseKmsAsymmetricSignatureAlgorithm(algo string) (kmsasymmetricsignature.AsymmetricSignatureAlgorithm, error) {
	val, ok := kmsasymmetricsignature.AsymmetricSignatureAlgorithm_value[algo]
	if !ok {
		return kmsasymmetricsignature.AsymmetricSignatureAlgorithm(0), fmt.Errorf("value for 'signature_algorithm' should be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(kmsasymmetricsignature.AsymmetricSignatureAlgorithm_value)), algo)
	}
	return kmsasymmetricsignature.AsymmetricSignatureAlgorithm(val), nil
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
		HostAffinityRules: expandHostAffinityRulesSpec(ruleSpecs),
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

	return &metadataOptions
}

func expandHostAffinityRulesSpec(ruleSpecs []interface{}) []*compute.PlacementPolicy_HostAffinityRule {
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

func expandLocalDiskSpecs(disks interface{}) []*compute.AttachedLocalDiskSpec {
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
			localDiskSpecs = append(localDiskSpecs, &compute.AttachedLocalDiskSpec{
				Size: diskSize,
			})
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
	}

	return []map[string]interface{}{metadataOptions}
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
		case *vpc.StaticRoute_GatewayId:
			m["gateway_id"] = h.GatewayId
		}

		staticRoutes.Add(m)
	}
	return staticRoutes
}

func expandStaticRoutes(v interface{}) ([]*vpc.StaticRoute, error) {
	staticRoutes := []*vpc.StaticRoute{}

	if v == nil {
		return staticRoutes, nil
	}

	routeList := v.(*schema.Set).List()
	for _, v := range routeList {
		sr, err := routeDescriptionToStaticRoute(v)
		if err != nil {
			return nil, fmt.Errorf("fail convert static route: %s", err)
		}
		staticRoutes = append(staticRoutes, sr)
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
	} else {
		return nil, errors.New("'static_route' should have a 'destination_prefix' field")
	}

	var nextHops = 0
	if v, ok := res["next_hop_address"].(string); ok && v != "" {
		sr.NextHop = &vpc.StaticRoute_NextHopAddress{
			NextHopAddress: v,
		}
		nextHops += 1
	}
	if v, ok := res["gateway_id"].(string); ok && v != "" {
		sr.NextHop = &vpc.StaticRoute_GatewayId{
			GatewayId: v,
		}
		nextHops += 1
	}

	if nextHops == 0 {
		return nil, errors.New("'static_route' should have a 'next_hop_address' or 'gateway_id' field")
	} else if nextHops > 1 {
		return nil, errors.New("'static_route' should have one of 'next_hop_address' or 'gateway_id' fields")
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

func securityRuleCidrsFromMap(res map[string]interface{}) (*vpc.CidrBlocks, bool) {
	var v4Blocks interface{} = nil
	var v6Blocks interface{} = nil

	if v, ok := res["v4_cidr_blocks"]; ok {
		v4Blocks = v
	}

	if v, ok := res["v6_cidr_blocks"]; ok {
		v6Blocks = v
	}

	return securityRuleCirds(v4Blocks, v6Blocks)
}

func securityRuleCidrsFromResourceData(data *schema.ResourceData) (*vpc.CidrBlocks, bool) {
	return securityRuleCirds(data.Get("v4_cidr_blocks"), data.Get("v6_cidr_blocks"))
}

func securityRuleCirds(v4Blocks, v6Blocks interface{}) (*vpc.CidrBlocks, bool) {
	cidr := new(vpc.CidrBlocks)

	if v4Blocks != nil {
		arr := v4Blocks.([]interface{})
		cidr.V4CidrBlocks = make([]string, len(arr))
		for i, c := range arr {
			cidr.V4CidrBlocks[i] = c.(string)
		}
	}

	if v6Blocks != nil {
		arr := v6Blocks.([]interface{})
		cidr.V6CidrBlocks = make([]string, len(arr))
		for i, c := range arr {
			cidr.V6CidrBlocks[i] = c.(string)
		}
	}

	ok := len(cidr.V4CidrBlocks) > 0 || len(cidr.V6CidrBlocks) > 0
	return cidr, ok
}

// return nil on ANY-typed port range
func securityRulePortsFromMap(res map[string]interface{}) (*vpc.PortRange, error) {
	port := int64(res["port"].(int))
	fromPort := int64(res["from_port"].(int))
	toPort := int64(res["to_port"].(int))

	return securityRulePorts(port, fromPort, toPort)
}

func securityRulePortsFromResourceData(data *schema.ResourceData) (*vpc.PortRange, error) {
	port := int64(data.Get("port").(int))
	fromPort := int64(data.Get("from_port").(int))
	toPort := int64(data.Get("to_port").(int))

	return securityRulePorts(port, fromPort, toPort)
}

func securityRulePorts(port, fromPort, toPort int64) (*vpc.PortRange, error) {
	if port == -1 && fromPort == -1 && toPort == -1 {
		return nil, nil
	}

	if port != -1 {
		if fromPort != -1 || toPort != -1 {
			return nil, fmt.Errorf("cannot set from_port/to_port with port")
		}
		fromPort = port
		toPort = port
	} else if fromPort == -1 || toPort == -1 {
		return nil, fmt.Errorf("port or from_port + to_port must be defined")
	}

	return &vpc.PortRange{FromPort: fromPort, ToPort: toPort}, nil
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

func expandDataprocCreateClusterConfigSpec(d *schema.ResourceData) (*dataproc.CreateClusterConfigSpec, error) {
	subclusters, err := expandDataprocSubclustersSpec(d)
	if err != nil {
		return nil, err
	}
	hadoop, err := expandDataprocHadoopConfig(d)
	if err != nil {
		return nil, err
	}
	return &dataproc.CreateClusterConfigSpec{
		VersionId:       d.Get("cluster_config.0.version_id").(string),
		Hadoop:          hadoop,
		SubclustersSpec: subclusters,
	}, nil
}

func expandDataprocHadoopConfig(d *schema.ResourceData) (*dataproc.HadoopConfig, error) {
	initActions, err := expandDataprocInitActions(d)
	if err != nil {
		return nil, err
	}
	return &dataproc.HadoopConfig{
		Services:              expandDataprocServices(d),
		Properties:            expandDataprocProperties(d),
		SshPublicKeys:         expandDataprocSSHPublicKeys(d),
		InitializationActions: initActions,
	}, nil
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

func expandDataprocInitActions(d *schema.ResourceData) ([]*dataproc.InitializationAction, error) {
	list := d.Get("cluster_config.0.hadoop.0.initialization_action").([]interface{})
	initActions := make([]*dataproc.InitializationAction, len(list))
	for index, element := range list {
		elementMap := element.(map[string]interface{})
		initAction := &dataproc.InitializationAction{
			Uri: elementMap["uri"].(string),
		}

		if elementMap["timeout"].(string) != "" {
			timeout, err := strconv.Atoi(elementMap["timeout"].(string))
			if err != nil {
				return nil, err
			}
			initAction.Timeout = int64(timeout)
		}

		argsList := elementMap["args"].([]interface{})
		if argsList != nil {
			args := make([]string, len(argsList))
			for listIndex, arg := range argsList {
				args[listIndex] = arg.(string)
			}
			initAction.Args = args
		}

		initActions[index] = initAction
	}
	return initActions, nil
}

func expandDataprocSubclustersSpec(d *schema.ResourceData) ([]*dataproc.CreateSubclusterConfigSpec, error) {
	rootKey := "cluster_config.0.subcluster_spec"
	list := d.Get(rootKey).([]interface{})
	subclusters := make([]*dataproc.CreateSubclusterConfigSpec, len(list))
	for index, element := range list {
		subcluster, err := expandDataprocSubclusterSpec(element)
		if err != nil {
			return nil, err
		}
		subclusters[index] = subcluster
	}

	return subclusters, nil
}

func expandDataprocSubclusterSpec(element interface{}) (*dataproc.CreateSubclusterConfigSpec, error) {
	subclusterSpec := element.(map[string]interface{})
	roleName := subclusterSpec["role"].(string)
	roleID := dataproc.Role_value[roleName]
	resourcesSpec := subclusterSpec["resources"].([]interface{})[0]

	subcluster := &dataproc.CreateSubclusterConfigSpec{
		Role:           dataproc.Role(roleID),
		Name:           subclusterSpec["name"].(string),
		SubnetId:       subclusterSpec["subnet_id"].(string),
		HostsCount:     int64(subclusterSpec["hosts_count"].(int)),
		Resources:      expandDataprocResources(resourcesSpec),
		AssignPublicIp: subclusterSpec["assign_public_ip"].(bool),
	}
	if v, ok := subclusterSpec["autoscaling_config"]; ok {
		autoscalingConfigs := v.([]interface{})
		if len(autoscalingConfigs) > 0 {
			autoscalingConfig, err := expandDataprocAutoscalingConfig(autoscalingConfigs[0])
			if err != nil {
				return nil, err
			}
			subcluster.AutoscalingConfig = autoscalingConfig
		}
	}

	return subcluster, nil
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

func stringToDuration(s string) (*duration.Duration, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &duration.Duration{Seconds: int64(i)}, nil
}

func expandAutoscalingDurationField(autoscalingConfigMap map[string]interface{}, fieldName string) (*duration.Duration, error) {
	if v, ok := autoscalingConfigMap[fieldName]; ok {
		durationSeconds := v.(string)
		if durationSeconds != "" {
			return stringToDuration(durationSeconds)
		}
	}
	return nil, nil
}

func expandDataprocAutoscalingConfig(r interface{}) (*dataproc.AutoscalingConfig, error) {
	autoscalingConfig := &dataproc.AutoscalingConfig{}
	autoscalingConfigMap := r.(map[string]interface{})
	log.Printf("[DEBUG] autoscalingConfigMap = %v", autoscalingConfigMap)
	if v, ok := autoscalingConfigMap["max_hosts_count"]; ok {
		if v.(int) >= 0 {
			autoscalingConfig.MaxHostsCount = int64(v.(int))
		}
	}
	if v, ok := autoscalingConfigMap["preemptible"]; ok {
		autoscalingConfig.Preemptible = v.(bool)
	}

	durationValue, err := expandAutoscalingDurationField(autoscalingConfigMap, "measurement_duration")
	if err != nil {
		return nil, err
	}
	autoscalingConfig.MeasurementDuration = durationValue

	durationValue, err = expandAutoscalingDurationField(autoscalingConfigMap, "warmup_duration")
	if err != nil {
		return nil, err
	}
	autoscalingConfig.WarmupDuration = durationValue

	durationValue, err = expandAutoscalingDurationField(autoscalingConfigMap, "stabilization_duration")
	if err != nil {
		return nil, err
	}
	autoscalingConfig.StabilizationDuration = durationValue

	if v, ok := autoscalingConfigMap["cpu_utilization_target"]; ok {
		value := v.(string)
		if value != "" {
			valueInt, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}

			autoscalingConfig.CpuUtilizationTarget = float64(valueInt)
		}
	}
	if v, ok := autoscalingConfigMap["decommission_timeout"]; ok {
		durationSeconds := v.(string)
		if durationSeconds != "" {
			durationSecondsInt, err := strconv.Atoi(durationSeconds)
			if err != nil {
				return nil, err
			}

			autoscalingConfig.DecommissionTimeout = int64(durationSecondsInt)
		}
	}

	return autoscalingConfig, nil
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
			"services":              flattenDataprocServices(config.Services),
			"properties":            config.Properties,
			"ssh_public_keys":       config.SshPublicKeys,
			"initialization_action": flattenInitActions(config.InitializationActions),
		},
	}
}

func flattenInitActions(actions []*dataproc.InitializationAction) []interface{} {
	result := make([]interface{}, len(actions))
	for index, action := range actions {
		result[index] = map[string]interface{}{
			"uri":     action.Uri,
			"args":    action.Args,
			"timeout": strconv.Itoa(int(action.Timeout)),
		}
	}
	return result
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
	result := map[string]interface{}{
		"id":               subcluster.Id,
		"name":             subcluster.Name,
		"role":             subcluster.Role.String(),
		"resources":        flattenDataprocResources(subcluster.Resources),
		"subnet_id":        subcluster.SubnetId,
		"hosts_count":      subcluster.HostsCount,
		"assign_public_ip": subcluster.AssignPublicIp,
	}
	if subcluster.AutoscalingConfig != nil {
		result["autoscaling_config"] = flattenDataprocAutoscalingConfig(subcluster.AutoscalingConfig)
	}
	return result
}

func flattenDataprocResources(r *dataproc.Resources) []map[string]interface{} {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return []map[string]interface{}{res}
}

func flattenDataprocAutoscalingConfig(r *dataproc.AutoscalingConfig) []map[string]interface{} {
	res := map[string]interface{}{}

	res["max_hosts_count"] = int(r.MaxHostsCount)
	res["preemptible"] = r.Preemptible
	res["measurement_duration"] = strconv.Itoa(int(r.MeasurementDuration.Seconds))
	res["warmup_duration"] = strconv.Itoa(int(r.WarmupDuration.Seconds))
	res["stabilization_duration"] = strconv.Itoa(int(r.StabilizationDuration.Seconds))
	res["cpu_utilization_target"] = strconv.FormatFloat(r.CpuUtilizationTarget, 'f', 0, 64)
	res["decommission_timeout"] = strconv.Itoa(int(r.DecommissionTimeout))

	return []map[string]interface{}{res}
}

func parseDataprocEnv(e string) (dataproc.Cluster_Environment, error) {
	v, ok := dataproc.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(dataproc.Cluster_Environment_value)), e)
	}
	return dataproc.Cluster_Environment(v), nil
}

func flattenSecurityGroupRulesProto(g *vpc.SecurityGroupRule) (port, fromPort, toPort int64) {
	port = -1
	fromPort = -1
	toPort = -1

	if ports := g.GetPorts(); ports != nil {
		if ports.FromPort == ports.ToPort {
			port = ports.FromPort
		} else {
			fromPort = ports.FromPort
			toPort = ports.ToPort
		}
	}

	return
}

func flattenSecurityGroupRulesSpec(sg []*vpc.SecurityGroupRule) (*schema.Set, *schema.Set) {
	ingress := schema.NewSet(resourceYandexVPCSecurityGroupRuleHash, nil)
	egress := schema.NewSet(resourceYandexVPCSecurityGroupRuleHash, nil)

	for _, g := range sg {
		r := make(map[string]interface{})

		r["id"] = g.GetId()
		r["description"] = g.GetDescription()
		r["labels"] = g.GetLabels()
		r["protocol"] = g.GetProtocolName()
		r["security_group_id"] = g.GetSecurityGroupId()
		r["predefined_target"] = g.GetPredefinedTarget()

		r["port"], r["from_port"], r["to_port"] = flattenSecurityGroupRulesProto(g)

		if cidr := g.GetCidrBlocks(); cidr != nil {
			if cidr.V4CidrBlocks != nil {
				r["v4_cidr_blocks"] = convertStringArrToInterface(cidr.V4CidrBlocks)
			}

			if cidr.V6CidrBlocks != nil {
				r["v6_cidr_blocks"] = convertStringArrToInterface(cidr.V6CidrBlocks)
			}
		}

		switch g.GetDirection().String() {
		case "INGRESS":
			ingress.Add(r)
		case "EGRESS":
			egress.Add(r)
		}
	}
	return ingress, egress
}

func flattenVpcAddressDnsRecords(specs []*vpc.DnsRecord) []map[string]interface{} {
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

func flattenExternalIpV4AddressSpec(address *vpc.ExternalIpv4Address) []interface{} {
	if address == nil {
		return nil
	}

	m := make(map[string]interface{})

	if address.Address != "" {
		m["address"] = address.GetAddress()
	}
	if address.ZoneId != "" {
		m["zone_id"] = address.GetZoneId()
	}

	if r := address.GetRequirements(); r != nil {
		if r.DdosProtectionProvider != "" {
			m["ddos_protection_provider"] = r.DdosProtectionProvider
		}
		if r.OutgoingSmtpCapability != "" {
			m["outgoing_smtp_capability"] = r.OutgoingSmtpCapability
		}
	}

	if len(m) > 0 {
		return []interface{}{m}
	}
	return nil
}

func expandAddressRequirements(addrDesc map[string]interface{}) (*vpc.AddressRequirements, bool) {
	var (
		set          bool
		requirements vpc.AddressRequirements
	)

	if v, ok := addrDesc["ddos_protection_provider"].(string); ok {
		set = true
		requirements.SetDdosProtectionProvider(v)
	}

	if v, ok := addrDesc["outgoing_smtp_capability"].(string); ok {
		set = true
		requirements.SetOutgoingSmtpCapability(v)
	}

	return &requirements, set
}

func expandVpcAddressDnsRecords(d *schema.ResourceData) ([]*vpc.DnsRecordSpec, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("dns_record"); !ok {
		return nil, nil
	}

	specs := v.([]interface{})
	recs := make([]*vpc.DnsRecordSpec, len(specs))

	for i, raw := range specs {
		d := raw.(map[string]interface{})
		r := &vpc.DnsRecordSpec{Fqdn: d["fqdn"].(string)}
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
	return recs, nil
}

func expandExternalIpv4Address(d *schema.ResourceData) (*vpc.ExternalIpv4AddressSpec, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("external_ipv4_address"); !ok {
		return nil, nil
	}

	addresses := v.([]interface{})

	if len(addresses) == 0 {
		return nil, nil
	}

	addrDesc, ok := addresses[0].(map[string]interface{})
	if !ok {
		return nil, addressError("fail to cast %#v to map[string]interface{}", addresses[0])
	}

	var addrSpec vpc.ExternalIpv4AddressSpec

	if v, ok := addrDesc["address"].(string); ok {
		addrSpec.SetAddress(v)
	}

	if v, ok := addrDesc["zone_id"].(string); ok {
		addrSpec.SetZoneId(v)
	}

	if requirements, ok := expandAddressRequirements(addrDesc); ok {
		addrSpec.SetRequirements(requirements)
	}

	return &addrSpec, nil
}

func expandAndValidateNetworkId(d *schema.ResourceData, config *Config) (string, error) {
	networkID := d.Get("network_id").(string)
	if config.Endpoint == common.DefaultEndpoint && len(networkID) == 0 {
		return "", fmt.Errorf("empty network_id field")
	}
	return networkID, nil
}

func flattenLocalDisks(instance *compute.Instance) []interface{} {
	if len(instance.LocalDisks) == 0 {
		return nil
	}
	result := make([]interface{}, len(instance.LocalDisks))
	for i, disk := range instance.LocalDisks {
		result[i] = map[string]interface{}{
			"size_bytes":  int(disk.Size),
			"device_name": disk.DeviceName,
		}
	}
	return result
}

func flattenSharedEgressGateway(sharedEgressGateway *vpc.SharedEgressGateway) []interface{} {
	if sharedEgressGateway == nil {
		return nil
	}

	m := make(map[string]interface{})

	return []interface{}{m}
}

func expandContainerRepositoryLifecyclePolicyRules(d *schema.ResourceData) ([]*containerregistry.LifecycleRule, error) {
	var rules []*containerregistry.LifecycleRule

	for _, key := range IterateKeys(d, "rule") {
		rule := &containerregistry.LifecycleRule{}

		if description, ok := d.GetOk(key + "description"); ok {
			rule.SetDescription(description.(string))
		}

		if expirePeriod, ok := d.GetOk(key + "expire_period"); ok {
			duration, err := parseDuration(expirePeriod.(string))
			if err != nil {
				return nil, err
			}

			rule.SetExpirePeriod(duration)
		}

		if retainedTop, ok := d.GetOk(key + "retained_top"); ok {
			rule.SetRetainedTop(int64(retainedTop.(int)))
		}

		if tagRegexp, ok := d.GetOk(key + "tag_regexp"); ok {
			rule.SetTagRegexp(tagRegexp.(string))
		}

		if untagged, ok := d.GetOk(key + "untagged"); ok {
			rule.SetUntagged(untagged.(bool))
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func flattenContainerRepositoryLifecyclePolicyRules(lifecycleRules []*containerregistry.LifecycleRule) []interface{} {
	s := make([]interface{}, 0, len(lifecycleRules))

	for _, lifecycleRule := range lifecycleRules {
		rule := flattenContainerRepositoryLifecyclePolicyRule(lifecycleRule)

		if len(rule) != 0 {
			s = append(s, rule)
		}
	}

	return s
}

func flattenContainerRepositoryLifecyclePolicyRule(lifecycleRule *containerregistry.LifecycleRule) map[string]interface{} {
	m := make(map[string]interface{})

	m["description"] = lifecycleRule.GetDescription()
	m["expire_period"] = formatDuration(lifecycleRule.GetExpirePeriod())
	m["tag_regexp"] = lifecycleRule.GetTagRegexp()
	m["untagged"] = lifecycleRule.GetUntagged()
	m["retained_top"] = lifecycleRule.GetRetainedTop()

	return m
}

func flattenSnapshotScheduleSchedulePolicy(policy *compute.SchedulePolicy) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"expression": policy.GetExpression(),
		"start_at":   getTimestamp(policy.GetStartAt()),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func expandSnapshotScheduleSchedulePolicy(d *schema.ResourceData) (*compute.SchedulePolicy, error) {
	val := new(compute.SchedulePolicy)

	if v, ok := d.GetOk("schedule_policy.0.start_at"); ok {
		startAt, err := parseTimestamp(v.(string))
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

func flattenSnapshotScheduleSnapshotSpec(spec *compute.SnapshotSpec) ([]map[string]interface{}, error) {
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
		labels, err := expandLabels(v.(map[string]interface{}))
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

func expandLoadtestingAgentLogSettingsParams(d *schema.ResourceData) (*ltagent.LogSettings, error) {
	var logGroupId string
	prefix := "log_settings.0."

	if v, ok := d.GetOk(prefix + "log_group_id"); ok {
		logGroupId = v.(string)
	}

	settings := &ltagent.LogSettings{
		CloudLogGroupId: logGroupId,
	}

	return settings, nil
}

func flattenLoadtestingAgentLogSettingsParams(agent *ltagent.Agent) ([]map[string]interface{}, error) {
	logSettings := agent.GetLogSettings()
	if logSettings == nil {
		return nil, nil
	}

	templateMap := make(map[string]interface{})

	templateMap["log_group_id"] = logSettings.GetCloudLogGroupId()

	return []map[string]interface{}{templateMap}, nil
}

func expandLoadtestingComputeInstanceTemplate(d *schema.ResourceData, config *Config) (*ltagent.CreateComputeInstance, error) {
	var zoneId, serviceAccountId, platformId string
	prefix := "compute_instance.0."

	if v, ok := d.GetOk(prefix + "zone_id"); ok {
		zoneId = v.(string)
	} else if config.Zone != "" {
		zoneId = config.Zone
	} else {
		return nil, fmt.Errorf("cannot determine zone: please set 'compute_instance.0.zone_id' key in this resource or at provider level")
	}
	if v, ok := d.GetOk(prefix + "service_account_id"); ok {
		serviceAccountId = v.(string)
	}
	if v, ok := d.GetOk(prefix + "platform_id"); ok {
		platformId = v.(string)
	}

	resourceSpec, err := expandLoadtestingComputeInstanceResourcesSpec(d, prefix)
	if err != nil {
		return nil, fmt.Errorf("Error create 'resources' object of api request: %s", err)
	}

	bootDiskSpec, err := expandLoadtestingComputeInstanceBootDiskSpec(d, prefix)
	if err != nil {
		return nil, fmt.Errorf("Error create 'boot_disk' object of api request: %s", err)
	}

	nicSpecs, err := expandLoadtestingComputeInstanceNetworkInterfaceSpecs(d, prefix)
	if err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	labels, err := expandLabels(d.Get(prefix + "labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance group: %s", err)
	}

	metadata, err := expandLabels(d.Get(prefix + "metadata"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding metadata while creating instance group: %s", err)
	}

	template := &ltagent.CreateComputeInstance{
		Labels:                labels,
		ZoneId:                zoneId,
		ResourcesSpec:         resourceSpec,
		Metadata:              metadata,
		BootDiskSpec:          bootDiskSpec,
		NetworkInterfaceSpecs: nicSpecs,
		ServiceAccountId:      serviceAccountId,
		PlatformId:            platformId,
	}

	return template, nil
}

func expandLoadtestingComputeInstanceResourcesSpec(d *schema.ResourceData, prefix string) (*compute.ResourcesSpec, error) {
	rs := &compute.ResourcesSpec{}
	prefix = prefix + "resources.0."

	if v, ok := d.GetOk(prefix + "cores"); ok {
		rs.Cores = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + "core_fraction"); ok {
		rs.CoreFraction = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + "memory"); ok {
		rs.Memory = toBytesFromFloat(v.(float64))
	}

	return rs, nil
}

func expandLoadtestingComputeInstanceBootDiskSpec(d *schema.ResourceData, prefix string) (*compute.AttachedDiskSpec, error) {
	ads := &compute.AttachedDiskSpec{}
	prefix = prefix + "boot_disk.0."

	if v, ok := d.GetOk(prefix + "auto_delete"); ok {
		ads.AutoDelete = v.(bool)
	}

	if v, ok := d.GetOk(prefix + "device_name"); ok {
		ads.DeviceName = v.(string)
	}

	if _, ok := d.GetOk(prefix + "initialize_params"); ok {
		bootDiskSpec, err := expandLoadtestingComputeInstanceDiskInitializeParamsSpec(d, prefix)
		if err != nil {
			return nil, err
		}
		ads.Disk = &compute.AttachedDiskSpec_DiskSpec_{
			DiskSpec: bootDiskSpec,
		}
	}

	return ads, nil
}

func expandLoadtestingComputeInstanceDiskInitializeParamsSpec(d *schema.ResourceData, prefix string) (*compute.AttachedDiskSpec_DiskSpec, error) {
	diskSpec := &compute.AttachedDiskSpec_DiskSpec{}
	prefix = prefix + "initialize_params.0."

	if v, ok := d.GetOk(prefix + "name"); ok {
		diskSpec.Name = v.(string)
	}

	if v, ok := d.GetOk(prefix + "description"); ok {
		diskSpec.Description = v.(string)
	}

	if v, ok := d.GetOk(prefix + "type"); ok {
		diskSpec.TypeId = v.(string)
	}

	if v, ok := d.GetOk(prefix + "size"); ok {
		diskSpec.Size = toBytes(v.(int))
	}

	if v, ok := d.GetOk(prefix + "block_size"); ok {
		diskSpec.BlockSize = int64(v.(int))
	}

	if diskSpec.Size == 0 {
		diskSpec.Size = toBytesFromFloat(15)
	}

	return diskSpec, nil
}

func expandLoadtestingComputeInstanceNetworkInterfaceSpecs(d *schema.ResourceData, prefix string) ([]*compute.NetworkInterfaceSpec, error) {
	nicsConfig := d.Get(prefix + "network_interface").([]interface{})
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

func flattenLoadtestingComputeInstanceTemplate(ctx context.Context, instance *compute.Instance, config *Config, origMetadata interface{}, origLabels interface{}) ([]map[string]interface{}, error) {
	templateMap := make(map[string]interface{})

	templateMap["zone_id"] = instance.GetZoneId()
	templateMap["labels"] = origLabels
	templateMap["computed_labels"] = instance.GetLabels()
	templateMap["metadata"] = origMetadata
	templateMap["computed_metadata"] = instance.GetMetadata()
	templateMap["service_account_id"] = instance.GetServiceAccountId()
	templateMap["platform_id"] = instance.GetPlatformId()

	resourceSpec, err := flattenLoadtestingComputeInstanceResources(instance)
	if err != nil {
		return nil, err
	}
	templateMap["resources"] = resourceSpec

	bootDiskSpec, err := flattenLoadtestingComputeInstanceBootDisk(ctx, instance, config.sdk.Compute().Disk())
	if err != nil {
		return []map[string]interface{}{templateMap}, err
	}
	templateMap["boot_disk"] = bootDiskSpec

	nics, err := flattenLoadtestingComputeInstanceNetworkInterfaces(instance)
	if err != nil {
		return []map[string]interface{}{templateMap}, err
	}
	templateMap["network_interface"] = nics

	return []map[string]interface{}{templateMap}, nil
}

func flattenLoadtestingComputeInstanceResources(instance *compute.Instance) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"cores":         int(instance.Resources.Cores),
		"core_fraction": int(instance.Resources.CoreFraction),
		"memory":        toGigabytesInFloat(instance.Resources.Memory),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func flattenLoadtestingComputeInstanceBootDisk(ctx context.Context, instance *compute.Instance, diskServiceClient ReducedDiskServiceClient) ([]map[string]interface{}, error) {
	attachedDisk := instance.GetBootDisk()
	if attachedDisk == nil {
		return nil, nil
	}

	bootDisk := map[string]interface{}{
		"auto_delete": attachedDisk.GetAutoDelete(),
		"device_name": attachedDisk.GetDeviceName(),
		"disk_id":     attachedDisk.GetDiskId(),
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
		"block_size":  int(disk.BlockSize),
		"type":        disk.TypeId,
	}}

	return []map[string]interface{}{bootDisk}, nil
}

func flattenLoadtestingComputeInstanceNetworkInterfaces(instance *compute.Instance) ([]map[string]interface{}, error) {
	nics := make([]map[string]interface{}, len(instance.NetworkInterfaces))

	for i, iface := range instance.NetworkInterfaces {
		index, err := strconv.Atoi(iface.Index)
		if err != nil {
			return nil, fmt.Errorf("Error while convert index of Network Interface: %s", err)
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

			if iface.PrimaryV4Address.OneToOneNat != nil {
				nics[i]["nat"] = true
				nics[i]["nat_ip_address"] = iface.PrimaryV4Address.OneToOneNat.Address
				nics[i]["nat_ip_version"] = iface.PrimaryV4Address.OneToOneNat.IpVersion.String()
			} else {
				nics[i]["nat"] = false
			}
		}

		if iface.PrimaryV6Address != nil {
			nics[i]["ipv6"] = true
			nics[i]["ipv6_address"] = iface.PrimaryV6Address.Address
		}
	}

	return nics, nil
}

func flattenUserSshKey(context context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userSSHKeyID := d.Id()
	res, ok := d.GetOk("user_ssh_key_id")
	if ok {
		userSSHKeyID = res.(string)
	}

	userSshKey, err := config.sdk.OrganizationManager().UserSshKey().Get(context,
		&organizationmanager.GetUserSshKeyRequest{
			UserSshKeyId: userSSHKeyID,
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("User ssh key %q", userSSHKeyID))
	}

	d.SetId(userSSHKeyID)
	_ = d.Set("subject_id", userSshKey.SubjectId)
	_ = d.Set("data", userSshKey.Data)
	_ = d.Set("name", userSshKey.Name)
	_ = d.Set("fingerprint", userSshKey.Fingerprint)
	_ = d.Set("organization_id", userSshKey.OrganizationId)
	_ = d.Set("created_at", getTimestamp(userSshKey.CreatedAt))
	_ = d.Set("expires_at", getTimestamp(userSshKey.ExpiresAt))

	return nil
}

func flattenUserSshKeySettings(v *organizationmanager.UserSshKeySettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["enabled"] = v.Enabled
	m["allow_manage_own_keys"] = v.AllowManageOwnKeys

	return []map[string]interface{}{m}, nil
}

func flattenSshCertificateSettings(v *organizationmanager.SshCertificateSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["enabled"] = v.Enabled

	return []map[string]interface{}{m}, nil
}

func expandUserSshKeySettings(d *schema.ResourceData) (*organizationmanager.UpdateOsLoginSettingsRequest_UserSshKeySettings, error) {
	val := new(organizationmanager.UpdateOsLoginSettingsRequest_UserSshKeySettings)

	if v, ok := d.GetOk("user_ssh_key_settings.0.enabled"); ok {
		val.SetEnabled(v.(bool))
	}

	if v, ok := d.GetOk("user_ssh_key_settings.0.allow_manage_own_keys"); ok {
		val.SetAllowManageOwnKeys(v.(bool))
	}

	return val, nil
}

func expandSshCertificateSettings(d *schema.ResourceData) (*organizationmanager.UpdateOsLoginSettingsRequest_SshCertificateSettings, error) {
	val := new(organizationmanager.UpdateOsLoginSettingsRequest_SshCertificateSettings)

	if v, ok := d.GetOk("ssh_certificate_settings.0.enabled"); ok {
		val.SetEnabled(v.(bool))
	}

	return val, nil
}

func flattenOsLoginSettings(context context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	organizationID := d.Get("organization_id").(string)

	osLoginSettings, err := config.sdk.OrganizationManager().OsLogin().GetSettings(context,
		&organizationmanager.GetOsLoginSettingsRequest{
			OrganizationId: organizationID,
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("OsLoginSettings %q", organizationID))
	}

	userSSHKeySettings, err := flattenUserSshKeySettings(osLoginSettings.UserSshKeySettings)
	if err != nil {
		return err
	}

	sshCertificateSettings, err := flattenSshCertificateSettings(osLoginSettings.SshCertificateSettings)
	if err != nil {
		return err
	}

	if err := d.Set("user_ssh_key_settings", userSSHKeySettings); err != nil {
		return err
	}
	d.SetId(organizationID)

	return d.Set("ssh_certificate_settings", sshCertificateSettings)
}
