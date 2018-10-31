package yandex

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/mitchellh/hashstructure"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/operation"
)

const (
	yandexComputeInstanceDefaultTimeout       = 5 * time.Minute
	yandexComputeInstanceDiskOperationTimeout = 1 * time.Minute
)

func resourceYandexComputeInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputeInstanceCreate,
		Read:   resourceYandexComputeInstanceRead,
		Update: resourceYandexComputeInstanceUpdate,
		Delete: resourceYandexComputeInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeInstanceDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeInstanceDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeInstanceDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "standard-v1",
			},
			"instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"memory": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"cores": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"core_fraction": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  100,
						},
					},
				},
			},
			"boot_disk": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_delete": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"disk_id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"boot_disk.initialize_params"},
						},
						"initialize_params": {
							Type:          schema.TypeList,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							MaxItems:      1,
							ConflictsWith: []string{"boot_disk.disk_id"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"description": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"size": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
										Default:      8,
									},
									"type_id": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  "network-hdd",
									},
									"image_id": {
										Type:          schema.TypeString,
										Optional:      true,
										Computed:      true,
										ForceNew:      true,
										ConflictsWith: []string{"boot_disk.initialize_params.snapshot_id"},
									},
									"snapshot_id": {
										Type:          schema.TypeString,
										Optional:      true,
										Computed:      true,
										ForceNew:      true,
										ConflictsWith: []string{"boot_disk.initialize_params.image_id"},
									},
								},
							},
						},
					},
				},
			},
			"network_interface": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"ipv6": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"ipv6_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"nat": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"nat_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"nat_ip_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"allow_stopping_for_update": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"secondary_disk": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_delete": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"device_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "READ_WRITE",
							ValidateFunc: validation.StringInSlice([]string{"READ_WRITE", "READ_ONLY"}, false),
						},
						"disk_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexComputeInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateInstanceRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error create instance: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error create instance: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	instance, ok := resp.(*compute.Instance)
	if !ok {
		return fmt.Errorf("response doesn't contain Instance")
	}

	d.SetId(instance.Id)

	return resourceYandexComputeInstanceRead(d, meta)
}

func resourceYandexComputeInstanceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	instance, err := config.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: d.Id(),
		View:       compute.InstanceView_FULL,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q", d.Get("name").(string)))
	}

	d.Set("metadata", instance.Metadata)
	d.Set("labels", instance.Labels)
	d.Set("platform_id", instance.PlatformId)
	d.Set("folder_id", instance.FolderId)
	d.Set("zone", instance.ZoneId)
	d.Set("name", instance.Name)
	d.Set("fqdn", instance.Fqdn)
	d.Set("description", instance.Description)
	d.Set("instance_id", instance.Id)

	if err := readInstanceResources(d, instance); err != nil {
		return err
	}

	if err := flattenBootDisk(d, meta, instance); err != nil {
		return err
	}

	disks, err := flattenSecondaryDisks(d, config, instance)
	if err != nil {
		return err
	}

	if err := d.Set("secondary_disk", disks); err != nil {
		return err
	}

	networkInterfaces, externalIP, err := flattenNetworkInterfaces(d, config, instance.NetworkInterfaces)
	if err != nil {
		return err
	}
	if err := d.Set("network_interface", networkInterfaces); err != nil {
		return err
	}

	if externalIP != "" {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": externalIP,
		})
	}

	return nil
}

func readInstanceResources(d *schema.ResourceData, instance *compute.Instance) error {
	resourceMap := map[string]interface{}{
		"cores":         int(instance.Resources.Cores),
		"core_fraction": int(instance.Resources.CoreFraction),
		"memory":        toGigabytes(instance.Resources.Memory),
	}

	err := d.Set("resources", []map[string]interface{}{resourceMap})
	return err
}

func flattenBootDisk(d *schema.ResourceData, meta interface{}, instance *compute.Instance) error {
	config := meta.(*Config)

	resourceMap := map[string]interface{}{
		"auto_delete": instance.BootDisk.AutoDelete,
		"device_name": instance.BootDisk.DeviceName,
		"disk_id":     instance.BootDisk.DiskId,
		"mode":        compute.AttachedDisk_Mode_name[int32(instance.BootDisk.Mode)],
	}

	disk, err := config.sdk.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
		DiskId: instance.BootDisk.DiskId,
	})
	if err != nil {
		return err
	}

	if _, ok := d.GetOk("boot_disk.0.initialize_params.#"); ok {
		m := d.Get("boot_disk.0.initialize_params")
		resourceMap["initialize_params"] = m
	} else {
		resourceMap["initialize_params"] = []map[string]interface{}{{
			"type_id":  disk.TypeId,
			"image_id": disk.GetSourceImageId(),
			"size":     toGigabytes(disk.Size),
		}}
	}

	return d.Set("boot_disk", []map[string]interface{}{resourceMap})
}

// revive:disable:var-naming
func flattenSecondaryDisks(d *schema.ResourceData, config *Config, instance *compute.Instance) ([]map[string]interface{}, error) {
	secondaryDisksInState := make(map[string]int)
	for i, v := range d.Get("secondary_disk").([]interface{}) {
		if v == nil {
			continue
		}
		disk := v.(map[string]interface{})
		diskId := disk["disk_id"].(string)
		secondaryDisksInState[diskId] = i
	}

	secondaryDisks := make([]map[string]interface{}, d.Get("secondary_disk.#").(int))

	for _, instanceDisk := range instance.SecondaryDisks {
		sdIndex, inState := secondaryDisksInState[instanceDisk.DiskId]
		disk := map[string]interface{}{
			"disk_id":     instanceDisk.DiskId,
			"device_name": instanceDisk.DeviceName,
			"mode":        compute.AttachedDisk_Mode_name[int32(instanceDisk.Mode)],
			"auto_delete": instanceDisk.AutoDelete,
		}
		if inState {
			secondaryDisks[sdIndex] = disk
		} else {
			secondaryDisks = append(secondaryDisks, disk)
		}
	}
	return secondaryDisks, nil
}

// revive:enable:var-naming

func flattenNetworkInterfaces(d *schema.ResourceData, config *Config, networkInterfaces []*compute.NetworkInterface) ([]map[string]interface{}, string, error) {
	flattened := make([]map[string]interface{}, len(networkInterfaces))
	var externalIP string

	for i, iface := range networkInterfaces {
		index, err := strconv.Atoi(iface.Index)
		if err != nil {
			return nil, "", fmt.Errorf("Error while convert index: %s", err)
		}

		flattened[i] = map[string]interface{}{
			"index":       index,
			"mac_address": iface.MacAddress,
			"subnet_id":   iface.SubnetId,
		}

		if iface.PrimaryV4Address != nil {
			flattened[i]["ip_address"] = iface.PrimaryV4Address.Address

			if iface.PrimaryV4Address.OneToOneNat != nil {
				flattened[i]["nat"] = true
				flattened[i]["nat_ip_address"] = iface.PrimaryV4Address.OneToOneNat.Address
				flattened[i]["nat_ip_version"] = iface.PrimaryV4Address.OneToOneNat.IpVersion.String()
				if externalIP == "" {
					externalIP = iface.PrimaryV4Address.OneToOneNat.Address
				}
			} else {
				flattened[i]["nat"] = false
			}
		}

		if iface.PrimaryV6Address != nil {
			flattened[i]["ipv6"] = true
			flattened[i]["ipv6_address"] = iface.PrimaryV6Address.Address
		}
	}

	return flattened, externalIP, nil
}

func resourceYandexComputeInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	instance, err := config.sdk.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
		InstanceId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q", d.Get("name").(string)))
	}

	d.Partial(true)

	labelPropName := "labels"
	if d.HasChange(labelPropName) {
		labelsProp, err := expandLabels(d.Get(labelPropName))
		if err != nil {
			return err
		}

		req := &compute.UpdateInstanceRequest{
			InstanceId: d.Id(),
			Labels:     labelsProp,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelPropName},
			},
		}

		err = makeInstanceUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(labelPropName)
	}

	metadataPropName := "metadata"
	if d.HasChange(metadataPropName) {
		metadataProp, err := expandLabels(d.Get(metadataPropName))
		if err != nil {
			return err
		}

		req := &compute.UpdateInstanceRequest{
			InstanceId: d.Id(),
			Metadata:   metadataProp,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{metadataPropName},
			},
		}

		err = makeInstanceUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(metadataPropName)
	}

	namePropName := "name"
	if d.HasChange(namePropName) {
		req := &compute.UpdateInstanceRequest{
			InstanceId: d.Id(),
			Name:       d.Get(namePropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{namePropName},
			},
		}

		err := makeInstanceUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(namePropName)
	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req := &compute.UpdateInstanceRequest{
			InstanceId: d.Id(),
			Name:       d.Get(descPropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{descPropName},
			},
		}

		err := makeInstanceUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(descPropName)
	}

	secDiskPropName := "secondary_disk"
	if d.HasChange(secDiskPropName) {
		if !d.Get("allow_stopping_for_update").(bool) {
			return fmt.Errorf("Changing the secondary_disk on an instance requires stopping it. " +
				"To acknowledge this action, please set allow_stopping_for_update = true in your config file.")
		}

		if err := makeInstanceActionRequest(instanceActionStop, d, meta); err != nil {
			return err
		}

		o, n := d.GetChange(secDiskPropName)

		// Keep track of disks currently in the instance. Because the yandex_compute_disk resource
		// can detach disks, it's possible that there are fewer disks currently attached than there
		// were at the time we ran terraform plan.
		currDisks := map[string]struct{}{}
		for _, disk := range instance.SecondaryDisks {
			currDisks[disk.DiskId] = struct{}{}
		}

		// Keep track of disks currently in state.
		// Since changing any field within the disk needs to detach+reattach it,
		// keep track of the hash of the disk spec.
		oDisks := map[uint64]string{}
		for _, disk := range o.([]interface{}) {
			diskConfig := disk.(map[string]interface{})
			diskSpec, err := prepareSecondaryDisk(diskConfig, d)
			if err != nil {
				return err
			}
			hash, err := hashstructure.Hash(*diskSpec, nil)
			if err != nil {
				return err
			}
			if _, ok := currDisks[diskSpec.GetDiskId()]; ok {
				oDisks[hash] = diskSpec.GetDiskId()
			}
		}

		// Keep track of new config's disks.
		// Since changing any field within the disk needs to detach+reattach it,
		// keep track of the hash of the full disk.
		// If a disk with a certain hash is only in the new config, it should be attached.
		nDisks := map[uint64]struct{}{}
		var attach []*compute.AttachedDiskSpec
		for _, disk := range n.([]interface{}) {
			diskConfig := disk.(map[string]interface{})
			diskSpec, err := prepareSecondaryDisk(diskConfig, d)
			if err != nil {
				return err
			}
			hash, err := hashstructure.Hash(*diskSpec, nil)
			if err != nil {
				return err
			}
			nDisks[hash] = struct{}{}

			if _, ok := oDisks[hash]; !ok {
				attach = append(attach, diskSpec)
			}
		}

		// If a source is only in the old config, it should be detached.
		// Detach the old disks.
		for hash, deviceID := range oDisks {
			if _, ok := nDisks[hash]; !ok {
				req := &compute.DetachInstanceDiskRequest{
					InstanceId: d.Id(),
					Disk: &compute.DetachInstanceDiskRequest_DiskId{
						DiskId: deviceID,
					},
				}

				err = makeDetachDiskRequest(req, d, meta)
				if err != nil {
					return err
				}
				log.Printf("[DEBUG] Successfully detached disk %s", deviceID)
			}
		}

		// Attach the new disks
		for _, diskSpec := range attach {
			req := &compute.AttachInstanceDiskRequest{
				InstanceId:       d.Id(),
				AttachedDiskSpec: diskSpec,
			}

			err := makeAttachDiskRequest(req, d, meta)
			if err != nil {
				return err
			}
			log.Printf("[DEBUG] Successfully attached disk %s", diskSpec.GetDiskId())
		}

		if err := makeInstanceActionRequest(instanceActionStart, d, meta); err != nil {
			return err
		}

		d.SetPartial(secDiskPropName)
	}

	d.Partial(false)

	return resourceYandexComputeInstanceRead(d, meta)
}

func resourceYandexComputeInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Instance %q", d.Id())

	req := &compute.DeleteInstanceRequest{
		InstanceId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Instance %q", d.Id())
	return nil
}

func prepareCreateInstanceRequest(d *schema.ResourceData, meta *Config) (*compute.CreateInstanceRequest, error) {
	zone, err := getZone(d, meta)
	if err != nil {
		return nil, err
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, err
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels: %s", err)
	}

	metadata, err := expandLabels(d.Get("metadata"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding metadata: %s", err)
	}

	req := &compute.CreateInstanceRequest{
		FolderId:    folderID,
		Hostname:    d.Get("hostname").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ZoneId:      zone,
		Labels:      labels,
		Metadata:    metadata,
		PlatformId:  d.Get("platform_id").(string),
	}

	if err := prepareResources(req, d); err != nil {
		return nil, fmt.Errorf("Error create 'resources_spec' object of api request: %s", err)
	}

	if err := prepareBootDisk(req, d); err != nil {
		return nil, fmt.Errorf("Error create 'boot_disk' object of api request: %s", err)
	}

	if err := prepareSecondaryDisks(req, d); err != nil {
		return nil, fmt.Errorf("Error create 'secondary_disk' object of api request: %s", err)
	}

	if err := prepareNetwork(req, d); err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	return req, nil
}

func prepareResources(req *compute.CreateInstanceRequest, d *schema.ResourceData) error {
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
		return fmt.Errorf("You should define 'resources' section for compute instance")
	}

	req.ResourcesSpec = rs
	return nil
}

func prepareSecondaryDisks(req *compute.CreateInstanceRequest, d *schema.ResourceData) error {
	secondaryDisksCount := d.Get("secondary_disk.#").(int)

	for i := 0; i < secondaryDisksCount; i++ {
		diskConfig := d.Get(fmt.Sprintf("secondary_disk.%d", i)).(map[string]interface{})

		disk, err := prepareSecondaryDisk(diskConfig, d)
		if err != nil {
			return err
		}
		req.SecondaryDiskSpecs = append(req.SecondaryDiskSpecs, disk)
	}
	return nil
}

func prepareSecondaryDisk(diskConfig map[string]interface{}, d *schema.ResourceData) (*compute.AttachedDiskSpec, error) {
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

func parseDiskMode(mode string) (compute.AttachedDiskSpec_Mode, error) {
	val, ok := compute.AttachedDiskSpec_Mode_value[mode]
	if !ok {
		return compute.AttachedDiskSpec_MODE_UNSPECIFIED, fmt.Errorf("value for 'mode' should be 'READ_WRITE' or 'READ_ONLY', not '%s'", mode)
	}
	return compute.AttachedDiskSpec_Mode(val), nil
}

func prepareBootDisk(req *compute.CreateInstanceRequest, d *schema.ResourceData) error {
	req.BootDiskSpec = new(compute.AttachedDiskSpec)

	if v, ok := d.GetOk("boot_disk.0.auto_delete"); ok {
		req.BootDiskSpec.AutoDelete = v.(bool)
	}

	if v, ok := d.GetOk("boot_disk.0.device_name"); ok {
		req.BootDiskSpec.DeviceName = v.(string)
	}

	if v, ok := d.GetOk("boot_disk.0.mode"); ok {
		diskMode, err := parseDiskMode(v.(string))
		if err != nil {
			return err
		}
		req.BootDiskSpec.Mode = diskMode
	}

	// use explicit disk
	if v, ok := d.GetOk("boot_disk.0.disk_id"); ok {
		req.BootDiskSpec.Disk = &compute.AttachedDiskSpec_DiskId{
			DiskId: v.(string),
		}
		return nil
	}

	// create new one disk
	if _, ok := d.GetOk("boot_disk.0.initialize_params"); ok {
		diskSpec, err := selectBootDiskSource(d)
		if err != nil {
			return err
		}

		req.BootDiskSpec.Disk = &compute.AttachedDiskSpec_DiskSpec_{
			DiskSpec: diskSpec,
		}
	}

	return nil
}

func selectBootDiskSource(d *schema.ResourceData) (*compute.AttachedDiskSpec_DiskSpec, error) {
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

func prepareNetwork(req *compute.CreateInstanceRequest, d *schema.ResourceData) error {
	nicsConfig := d.Get("network_interface").([]interface{})
	nics := make([]*compute.NetworkInterfaceSpec, len(nicsConfig))

	for i, raw := range nicsConfig {
		data := raw.(map[string]interface{})

		subnetID := data["subnet_id"].(string)
		if subnetID == "" {
			return fmt.Errorf("NIC number %d does not have a 'subnet_id' attribute defined", i)
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

	req.NetworkInterfaceSpecs = nics
	return nil
}

func makeInstanceUpdateRequest(req *compute.UpdateInstanceRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().Update(ctx, req))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Instance %q: %s", d.Id(), err)
	}

	return nil
}

func makeInstanceActionRequest(action instanceAction, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	instanceID := d.Id()
	var err error
	var op *operation.Operation

	log.Printf("[DEBUG] Prepare to run %s action on instance %s", action, instanceID)

	switch action {
	case instanceActionStop:
		{
			op, err = config.sdk.WrapOperation(config.sdk.Compute().Instance().
				Stop(ctx, &compute.StopInstanceRequest{
					InstanceId: instanceID,
				}))
		}
	case instanceActionStart:
		{
			op, err = config.sdk.WrapOperation(config.sdk.Compute().Instance().
				Start(ctx, &compute.StartInstanceRequest{
					InstanceId: instanceID,
				}))
		}
	default:
		return fmt.Errorf("Action %s not supported", action)
	}

	if err != nil {
		log.Printf("[DEBUG] Error while run %s action on instance %s: %s", action, instanceID, err)
		return fmt.Errorf("Error while run %s action on Instance %s: %s", action, instanceID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		log.Printf("[DEBUG] Error while wait %s action on instance %s: %s", action, instanceID, err)
		return fmt.Errorf("Error while wait %s action on Instance %s: %s", action, instanceID, err)
	}

	return nil
}

func makeDetachDiskRequest(req *compute.DetachInstanceDiskRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), yandexComputeInstanceDiskOperationTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().DetachDisk(ctx, req))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error detach Disk %s for Instance %q: %s", req.GetDiskId(), d.Id(), err)
	}

	return nil
}

func makeAttachDiskRequest(req *compute.AttachInstanceDiskRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), yandexComputeInstanceDiskOperationTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().AttachDisk(ctx, req))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error attach Disk %s for Instance %q: %s", req.AttachedDiskSpec.GetDiskId(), d.Id(), err)
	}

	return nil
}
