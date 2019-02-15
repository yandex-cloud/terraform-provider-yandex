package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
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

									"type": {
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
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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

						"nat": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"index": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"mac_address": {
							Type:     schema.TypeString,
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

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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

			"allow_stopping_for_update": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"secondary_disk": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_id": {
							Type:     schema.TypeString,
							Required: true,
						},

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
					},
				},
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
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
		return fmt.Errorf("Error while requesting API to create instance: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create instance: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return fmt.Errorf("Instance creation failed: %s", err)
	}

	instance, ok := resp.(*compute.Instance)
	if !ok {
		return fmt.Errorf("Create response doesn't contain Instance")
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

	resources, err := flattenInstanceResources(instance)
	if err != nil {
		return err
	}

	bootDisk, err := flattenInstanceBootDisk(instance, config.sdk.Compute().Disk())
	if err != nil {
		return err
	}

	secondaryDisks, err := flattenInstanceSecondaryDisks(instance)
	if err != nil {
		return err
	}

	networkInterfaces, externalIP, internalIP, err := flattenInstanceNetworkInterfaces(instance)
	if err != nil {
		return err
	}

	createdAt, err := getTimestamp(instance.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("platform_id", instance.PlatformId)
	d.Set("folder_id", instance.FolderId)
	d.Set("zone", instance.ZoneId)
	d.Set("name", instance.Name)
	d.Set("fqdn", instance.Fqdn)
	d.Set("description", instance.Description)
	d.Set("status", strings.ToLower(instance.Status.String()))

	if err := d.Set("metadata", instance.Metadata); err != nil {
		return err
	}

	if err := d.Set("labels", instance.Labels); err != nil {
		return err
	}

	if err := d.Set("resources", resources); err != nil {
		return err
	}

	if err := d.Set("boot_disk", bootDisk); err != nil {
		return err
	}

	if err := d.Set("secondary_disk", secondaryDisks); err != nil {
		return err
	}

	if err := d.Set("network_interface", networkInterfaces); err != nil {
		return err
	}

	connIP := externalIP
	if connIP == "" {
		connIP = internalIP
	}

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": connIP,
	})

	return nil
}

// revive:enable:var-naming

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
			diskSpec, err := expandSecondaryDiskSpec(diskConfig)
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
			diskSpec, err := expandSecondaryDiskSpec(diskConfig)
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
		return nil, fmt.Errorf("Error getting zone while creating instance: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating instance: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance: %s", err)
	}

	metadata, err := expandLabels(d.Get("metadata"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding metadata while creating instance: %s", err)
	}

	resourcesSpec, err := expandInstanceResourcesSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'resources_spec' object of api request: %s", err)
	}

	bootDiskSpec, err := expandInstanceBootDiskSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'boot_disk' object of api request: %s", err)
	}

	secondaryDiskSpecs, err := expandInstanceSecondaryDiskSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'secondary_disk' object of api request: %s", err)
	}

	nicSpecs, err := expandInstanceNetworkInterfaceSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	req := &compute.CreateInstanceRequest{
		FolderId:              folderID,
		Hostname:              d.Get("hostname").(string),
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		PlatformId:            d.Get("platform_id").(string),
		ZoneId:                zone,
		Labels:                labels,
		Metadata:              metadata,
		ResourcesSpec:         resourcesSpec,
		BootDiskSpec:          bootDiskSpec,
		SecondaryDiskSpecs:    secondaryDiskSpecs,
		NetworkInterfaceSpecs: nicSpecs,
	}

	return req, nil
}

func makeInstanceUpdateRequest(req *compute.UpdateInstanceRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Instance %q: %s", d.Id(), err)
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
		return fmt.Errorf("Error while requesting API to detach Disk %s from Instance %q: %s", req.GetDiskId(), d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error detach Disk %s from Instance %q: %s", req.GetDiskId(), d.Id(), err)
	}

	return nil
}

func makeAttachDiskRequest(req *compute.AttachInstanceDiskRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), yandexComputeInstanceDiskOperationTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().AttachDisk(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to attach Disk %s to Instance %q: %s", req.AttachedDiskSpec.GetDiskId(), d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error attach Disk %s to Instance %q: %s", req.AttachedDiskSpec.GetDiskId(), d.Id(), err)
	}

	return nil
}
