package yandex

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mitchellh/hashstructure"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/operation"
)

const (
	yandexComputeInstanceDefaultTimeout       = 5 * time.Minute
	yandexComputeInstanceDiskOperationTimeout = 1 * time.Minute
	yandexComputeInstanceDeallocationTimeout  = 15 * time.Second
	yandexComputeInstanceMoveTimeout          = 1 * time.Minute
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

		SchemaVersion: 1,

		MigrateState: resourceComputeInstanceMigrateState,

		Schema: map[string]*schema.Schema{
			"resources": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"memory": {
							Type:         schema.TypeFloat,
							Required:     true,
							ForceNew:     false,
							ValidateFunc: FloatAtLeast(0.0),
						},

						"cores": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: false,
						},

						"gpus": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},

						"core_fraction": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: false,
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
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},

									"block_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
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

			"network_acceleration_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "standard",
				ValidateFunc: validation.StringInSlice([]string{"standard", "software_accelerated"}, false),
			},

			"network_interface": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"ipv4": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
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
						},

						"nat": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
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
							Optional: true,
							Computed: true,
						},

						"nat_ip_version": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
						},

						"dns_record": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fqdn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"dns_zone_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"ptr": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},

						"ipv6_dns_record": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fqdn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"dns_zone_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"ptr": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},

						"nat_dns_record": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fqdn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"dns_zone_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"ptr": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				DiffSuppressFunc: hostnameDiffSuppressFunc,
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
				ForceNew: false,
				Default:  "standard-v1",
			},

			"allow_stopping_for_update": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"allow_recreate": {
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

			"scheduling_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"preemptible": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"placement_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"placement_group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"host_affinity_rules": {
							Type:       schema.TypeList,
							Computed:   true,
							Optional:   true,
							ConfigMode: schema.SchemaConfigModeAttr,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"op": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice(
											generateHostAffinityRuleOperators(), false),
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
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

			"local_disk": {
				Type:         schema.TypeList,
				Optional:     true,
				RequiredWith: []string{"placement_policy"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_bytes": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"device_name": {
							Type:     schema.TypeString,
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create instance: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get instance create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateInstanceMetadata)
	if !ok {
		return fmt.Errorf("could not get Instance ID from create operation metadata")
	}

	d.SetId(md.InstanceId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create instance: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Instance creation failed: %s", err)
	}

	return resourceYandexComputeInstanceRead(d, meta)
}

func resourceYandexComputeInstanceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
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

	bootDisk, err := flattenInstanceBootDisk(ctx, instance, config.sdk.Compute().Disk())
	if err != nil {
		return err
	}

	secondaryDisks, err := flattenInstanceSecondaryDisks(instance)
	if err != nil {
		return err
	}

	schedulingPolicy, err := flattenInstanceSchedulingPolicy(instance)
	if err != nil {
		return err
	}

	placementPolicy, err := flattenInstancePlacementPolicy(instance)
	if err != nil {
		return err
	}

	networkInterfaces, externalIP, internalIP, err := flattenInstanceNetworkInterfaces(instance)
	if err != nil {
		return err
	}

	localDisks := flattenLocalDisks(instance)

	d.Set("created_at", getTimestamp(instance.CreatedAt))
	d.Set("platform_id", instance.PlatformId)
	d.Set("folder_id", instance.FolderId)
	d.Set("zone", instance.ZoneId)
	d.Set("name", instance.Name)
	d.Set("fqdn", instance.Fqdn)
	d.Set("description", instance.Description)
	d.Set("service_account_id", instance.ServiceAccountId)
	d.Set("status", strings.ToLower(instance.Status.String()))

	hostname, err := parseHostnameFromFQDN(instance.Fqdn)
	if err != nil {
		return err
	}
	d.Set("hostname", hostname)

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

	if err := d.Set("scheduling_policy", schedulingPolicy); err != nil {
		return err
	}

	if err := d.Set("placement_policy", placementPolicy); err != nil {
		return err
	}

	if err := d.Set("local_disk", localDisks); err != nil {
		return err
	}

	if instance.NetworkSettings != nil {
		d.Set("network_acceleration_type", strings.ToLower(instance.NetworkSettings.Type.String()))
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

	ctx := config.Context()

	instance, err := config.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q", d.Get("name").(string)))
	}

	d.Partial(true)

	folderPropName := "folder_id"
	if d.HasChange(folderPropName) {
		if !d.Get("allow_recreate").(bool) {
			if err := ensureAllowStoppingForUpdate(d, folderPropName); err != nil {
				return err
			}

			if instance.Status != compute.Instance_STOPPED {
				if err := makeInstanceActionRequest(instanceActionStop, d, meta); err != nil {
					return err
				}
			}

			req := &compute.MoveInstanceRequest{
				InstanceId:          d.Id(),
				DestinationFolderId: d.Get(folderPropName).(string),
			}

			if err := makeInstanceMoveRequest(req, d, meta); err != nil {
				return err
			}

			if err := makeInstanceActionRequest(instanceActionStart, d, meta); err != nil {
				return err
			}

		} else {
			if err := resourceYandexComputeInstanceDelete(d, meta); err != nil {
				return err
			}
			if err := resourceYandexComputeInstanceCreate(d, meta); err != nil {
				return err
			}
		}
	}

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

	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req := &compute.UpdateInstanceRequest{
			InstanceId:  d.Id(),
			Description: d.Get(descPropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{descPropName},
			},
		}

		err := makeInstanceUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	serviceAccountPropName := "service_account_id"
	if d.HasChange(serviceAccountPropName) {
		req := &compute.UpdateInstanceRequest{
			InstanceId:       d.Id(),
			ServiceAccountId: d.Get(serviceAccountPropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{serviceAccountPropName},
			},
		}

		err := makeInstanceUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	networkInterfacesPropName := "network_interface"
	needUpdateInterfacesOnStoppedInstance := false
	var addNatRequests []*compute.AddInstanceOneToOneNatRequest
	var removeNatRequests []*compute.RemoveInstanceOneToOneNatRequest
	var updateInterfaceRequests []*compute.UpdateInstanceNetworkInterfaceRequest
	if d.HasChange(networkInterfacesPropName) {
		o, n := d.GetChange(networkInterfacesPropName)
		oldList := o.([]interface{})
		newList := n.([]interface{})

		if len(oldList) != len(newList) {
			return fmt.Errorf("Changing count of network interfaces is't supported yet")
		}

		for ifaceIndex := 0; ifaceIndex < len(oldList); ifaceIndex++ {
			log.Printf("[DEBUG] Processing interface #%d", ifaceIndex)
			oldIface := oldList[ifaceIndex].(map[string]interface{})
			newIface := newList[ifaceIndex].(map[string]interface{})
			req := &compute.UpdateInstanceNetworkInterfaceRequest{
				InstanceId:            d.Id(),
				NetworkInterfaceIndex: fmt.Sprint(ifaceIndex),
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{},
				},
			}

			oldV4Spec, err := expandPrimaryV4AddressSpec(oldIface)
			if err != nil {
				return err
			}
			oldV6Spec, err := expandPrimaryV6AddressSpec(oldIface)
			if err != nil {
				return err
			}
			newV4Spec, err := expandPrimaryV4AddressSpec(newIface)
			if err != nil {
				return err
			}
			newV6Spec, err := expandPrimaryV6AddressSpec(newIface)
			if err != nil {
				return err
			}

			if oldIface["subnet_id"].(string) != newIface["subnet_id"].(string) {
				// change subnet, update all the properties!
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, "subnet_id", "primary_v4_address_spec", "primary_v6_address_spec")
				// ...on stopped instance
				needUpdateInterfacesOnStoppedInstance = true

				req.SubnetId = newIface["subnet_id"].(string)
				req.PrimaryV4AddressSpec = newV4Spec
				req.PrimaryV6AddressSpec = newV6Spec
			} else {
				if wantChangeAddressSpec(oldV4Spec, newV4Spec) {
					// change primary v4 address
					req.UpdateMask.Paths = append(req.UpdateMask.Paths, "primary_v4_address_spec")
					// ...on stopped instance
					needUpdateInterfacesOnStoppedInstance = true

					req.PrimaryV4AddressSpec = newV4Spec
				} else {
					if wantChangeNatSpec(oldV4Spec.OneToOneNatSpec, newV4Spec.OneToOneNatSpec) {
						// changing nat address on maybe running instance, safer to use add/remove nat calls
						if oldV4Spec.OneToOneNatSpec != nil {
							removeNatRequests = append(removeNatRequests, &compute.RemoveInstanceOneToOneNatRequest{
								InstanceId:            d.Id(),
								NetworkInterfaceIndex: fmt.Sprint(ifaceIndex),
							})
						}
						if newV4Spec.OneToOneNatSpec != nil {
							addNatRequests = append(addNatRequests, &compute.AddInstanceOneToOneNatRequest{
								InstanceId:            d.Id(),
								NetworkInterfaceIndex: fmt.Sprint(ifaceIndex),
								OneToOneNatSpec:       newV4Spec.OneToOneNatSpec,
							})
						}
					}
				}

				if wantChangeAddressSpec(oldV6Spec, newV6Spec) {
					// change primary v6 address
					req.UpdateMask.Paths = append(req.UpdateMask.Paths, "primary_v6_address_spec")
					// ...on stopped instance
					needUpdateInterfacesOnStoppedInstance = true

					req.PrimaryV6AddressSpec = newV6Spec
				}
			}

			oldSgs := expandSecurityGroupIds(oldIface["security_group_ids"])
			newSgs := expandSecurityGroupIds(newIface["security_group_ids"])
			if !reflect.DeepEqual(oldSgs, newSgs) {
				log.Printf("[DEBUG]  changing sgs form %s to %s", oldSgs, newSgs)
				// change security groups
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")

				req.SecurityGroupIds = newSgs
			}

			if len(req.UpdateMask.Paths) > 0 {
				updateInterfaceRequests = append(updateInterfaceRequests, req)
			}
		}

		if !needUpdateInterfacesOnStoppedInstance && (len(removeNatRequests) > 0 || len(addNatRequests) > 0 || len(updateInterfaceRequests) > 0) {
			for _, req := range removeNatRequests {
				err := makeInstanceRemoveOneToOneNatRequest(req, d, meta)
				if err != nil {
					return err
				}
			}
			for _, req := range addNatRequests {
				err := makeInstanceAddOneToOneNatRequest(req, d, meta)
				if err != nil {
					return err
				}
			}
			for _, req := range updateInterfaceRequests {
				err := makeInstanceUpdateNetworkInterfaceRequest(req, d, meta)
				if err != nil {
					return err
				}
			}

		}
	}

	secDiskPropName := "secondary_disk"
	if d.HasChange(secDiskPropName) {
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
			hash, err := hashstructure.Hash(diskSpec, nil)
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
			hash, err := hashstructure.Hash(diskSpec, nil)
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

				err = makeDetachDiskRequest(req, meta)
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

			err := makeAttachDiskRequest(req, meta)
			if err != nil {
				return err
			}
			log.Printf("[DEBUG] Successfully attached disk %s", diskSpec.GetDiskId())
		}
	}

	resourcesPropName := "resources"
	platformIDPropName := "platform_id"
	networkAccelerationTypePropName := "network_acceleration_type"
	schedulingPolicyName := "scheduling_policy"
	placementPolicyPropName := "placement_policy"
	properties := []string{
		resourcesPropName,
		platformIDPropName,
		networkAccelerationTypePropName,
		schedulingPolicyName,
		placementPolicyPropName,
	}
	if d.HasChange(resourcesPropName) || d.HasChange(platformIDPropName) || d.HasChange(networkAccelerationTypePropName) ||
		needUpdateInterfacesOnStoppedInstance || d.HasChange(schedulingPolicyName) || d.HasChange(placementPolicyPropName) {
		if err := ensureAllowStoppingForUpdate(d, properties...); err != nil {
			return err
		}
		if err := makeInstanceActionRequest(instanceActionStop, d, meta); err != nil {
			return err
		}

		instanceStoppedAt := time.Now()

		// update platform, resources and network_settings in one request
		if d.HasChange(resourcesPropName) || d.HasChange(platformIDPropName) || d.HasChange(networkAccelerationTypePropName) ||
			d.HasChange(placementPolicyPropName) || d.HasChange(schedulingPolicyName) {
			req := &compute.UpdateInstanceRequest{
				InstanceId: d.Id(),
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{},
				},
			}

			if d.HasChange(resourcesPropName) {
				spec, err := expandInstanceResourcesSpec(d)
				if err != nil {
					return err
				}

				req.ResourcesSpec = spec
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, "resources_spec")
			}

			if d.HasChange(platformIDPropName) {
				req.PlatformId = d.Get(platformIDPropName).(string)
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, platformIDPropName)
			}

			if d.HasChange(networkAccelerationTypePropName) {
				networkSettings, err := expandInstanceNetworkSettingsSpecs(d)
				if err != nil {
					return err
				}

				req.NetworkSettings = networkSettings
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, "network_settings")
			}

			if d.HasChange(schedulingPolicyName) {
				schedulingPolicy, err := expandInstanceSchedulingPolicy(d)
				if err != nil {
					return err
				}

				req.SchedulingPolicy = schedulingPolicy
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, "scheduling_policy.preemptible")
			}

			if d.HasChange(placementPolicyPropName) {
				placementPolicy, paths := preparePlacementPolicyForUpdateRequest(d)
				req.PlacementPolicy = placementPolicy
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, paths...)
			}

			err = makeInstanceUpdateRequest(req, d, meta)
			if err != nil {
				return err
			}
		}

		// update interfaces on stopped instance
		if needUpdateInterfacesOnStoppedInstance {
			// wait for resource deallocation
			timeSinceInstanceStopped := time.Since(instanceStoppedAt)
			if timeSinceInstanceStopped < yandexComputeInstanceDeallocationTimeout {
				sleepTime := yandexComputeInstanceDeallocationTimeout - timeSinceInstanceStopped
				log.Printf("[DEBUG] Sleeping %s, waiting for deallocation", sleepTime)
				time.Sleep(sleepTime)
			}
			for _, req := range removeNatRequests {
				err := makeInstanceRemoveOneToOneNatRequest(req, d, meta)
				if err != nil {
					return err
				}
			}
			for _, req := range addNatRequests {
				err := makeInstanceAddOneToOneNatRequest(req, d, meta)
				if err != nil {
					return err
				}
			}
			for _, req := range updateInterfaceRequests {
				err := makeInstanceUpdateNetworkInterfaceRequest(req, d, meta)
				if err != nil {
					return err
				}
			}

		}

		if err := makeInstanceActionRequest(instanceActionStart, d, meta); err != nil {
			return err
		}
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
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

	bootDiskSpec, err := expandInstanceBootDiskSpec(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error create 'boot_disk' object of api request: %s", err)
	}

	secondaryDiskSpecs, err := expandInstanceSecondaryDiskSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'secondary_disk' object of api request: %s", err)
	}

	networkSettingsSpecs, err := expandInstanceNetworkSettingsSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	nicSpecs, err := expandInstanceNetworkInterfaceSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	schedulingPolicy, err := expandInstanceSchedulingPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'scheduling_policy' object of api request: %s", err)
	}

	placementPolicy, err := expandInstancePlacementPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error create 'placement_policy' object of api request: %s", err)
	}

	localDisks := expandLocalDiskSpecs(d.Get("local_disk"))

	req := &compute.CreateInstanceRequest{
		FolderId:              folderID,
		Hostname:              d.Get("hostname").(string),
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		PlatformId:            d.Get("platform_id").(string),
		ServiceAccountId:      d.Get("service_account_id").(string),
		ZoneId:                zone,
		Labels:                labels,
		Metadata:              metadata,
		ResourcesSpec:         resourcesSpec,
		BootDiskSpec:          bootDiskSpec,
		SecondaryDiskSpecs:    secondaryDiskSpecs,
		NetworkSettings:       networkSettingsSpecs,
		NetworkInterfaceSpecs: nicSpecs,
		SchedulingPolicy:      schedulingPolicy,
		PlacementPolicy:       placementPolicy,
		LocalDiskSpecs:        localDisks,
	}

	return req, nil
}

func parseHostnameFromFQDN(fqdn string) (string, error) {
	if !strings.Contains(fqdn, ".") {
		return fqdn + ".", nil
	}
	if strings.HasSuffix(fqdn, ".auto.internal") {
		return "", nil
	}
	if strings.HasSuffix(fqdn, ".internal") {
		p := strings.Split(fqdn, ".")
		return p[0], nil
	}

	return fqdn, nil
}

func wantChangeAddressSpec(old *compute.PrimaryAddressSpec, new *compute.PrimaryAddressSpec) bool {
	if old == nil && new == nil {
		return false
	}

	if (old != nil && new == nil) || (old == nil && new != nil) {
		return true
	}

	if new.Address != "" && old.Address != new.Address {
		return true
	}

	if len(old.DnsRecordSpecs) != len(new.DnsRecordSpecs) {
		return true
	}

	for i, oldrs := range old.DnsRecordSpecs {
		newrs := new.DnsRecordSpecs[i]
		if differentRecordSpec(oldrs, newrs) {
			return true
		}
	}
	return false
}

func wantChangeNatSpec(old *compute.OneToOneNatSpec, new *compute.OneToOneNatSpec) bool {
	if old == nil && new == nil {
		return false
	}

	if (old != nil && new == nil) || (old == nil && new != nil) {
		return true
	}

	if new.Address != "" && old.Address != new.Address {
		return true
	}

	if len(old.DnsRecordSpecs) != len(new.DnsRecordSpecs) {
		return true
	}

	for i, oldrs := range old.DnsRecordSpecs {
		newrs := new.DnsRecordSpecs[i]
		if differentRecordSpec(oldrs, newrs) {
			return true
		}
	}
	return false
}

func makeInstanceUpdateRequest(req *compute.UpdateInstanceRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
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

func makeInstanceUpdateNetworkInterfaceRequest(req *compute.UpdateInstanceNetworkInterfaceRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().UpdateNetworkInterface(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update network interface for Instance %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Instance %q: %s", d.Id(), err)
	}

	return nil
}

func makeInstanceAddOneToOneNatRequest(req *compute.AddInstanceOneToOneNatRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().AddOneToOneNat(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to add one-to-one nat for Instance %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Instance %q: %s", d.Id(), err)
	}

	return nil
}

func makeInstanceRemoveOneToOneNatRequest(req *compute.RemoveInstanceOneToOneNatRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().RemoveOneToOneNat(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to remove one-to-one nat for Instance %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Instance %q: %s", d.Id(), err)
	}

	return nil
}

func makeInstanceActionRequest(action instanceAction, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
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

func makeDetachDiskRequest(req *compute.DetachInstanceDiskRequest, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), yandexComputeInstanceDiskOperationTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().DetachDisk(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to detach Disk %s from Instance %q: %s", req.GetDiskId(), req.GetInstanceId(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error detach Disk %s from Instance %q: %s", req.GetDiskId(), req.GetInstanceId(), err)
	}

	return nil
}

func makeAttachDiskRequest(req *compute.AttachInstanceDiskRequest, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), yandexComputeInstanceDiskOperationTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().AttachDisk(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to attach Disk %s to Instance %q: %s", req.AttachedDiskSpec.GetDiskId(), req.GetInstanceId(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error attach Disk %s to Instance %q: %s", req.AttachedDiskSpec.GetDiskId(), req.GetInstanceId(), err)
	}

	return nil
}

func makeInstanceMoveRequest(req *compute.MoveInstanceRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), yandexComputeInstanceMoveTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Instance().Move(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to move Instance %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error moving Instance %q: %s", d.Id(), err)
	}

	return nil
}

func differentRecordSpec(r1, r2 *compute.DnsRecordSpec) bool {
	return r1.GetFqdn() != r2.GetFqdn() ||
		r1.GetDnsZoneId() != r2.GetDnsZoneId() ||
		r1.GetTtl() != r2.GetTtl() ||
		r1.GetPtr() != r2.GetPtr()
}

func generateHostAffinityRuleOperators() []string {
	operators := make([]string, 0, len(compute.PlacementPolicy_HostAffinityRule_Operator_value))
	for operatorName := range compute.PlacementPolicy_HostAffinityRule_Operator_value {
		operators = append(operators, operatorName)
	}
	return operators
}

func preparePlacementPolicyForUpdateRequest(d *schema.ResourceData) (*compute.PlacementPolicy, []string) {
	var placementPolicy compute.PlacementPolicy
	var paths []string
	if d.HasChange("placement_policy.0.placement_group_id") {
		placementPolicy.PlacementGroupId = d.Get("placement_policy.0.placement_group_id").(string)
		paths = append(paths, "placement_policy.placement_group_id")
	}

	if d.HasChange("placement_policy.0.host_affinity_rules") {
		rules := d.Get("placement_policy.0.host_affinity_rules").([]interface{})
		placementPolicy.HostAffinityRules = expandHostAffinityRulesSpec(rules)
		paths = append(paths, "placement_policy.host_affinity_rules")
	}
	return &placementPolicy, paths
}

func ensureAllowStoppingForUpdate(d *schema.ResourceData, propNames ...string) error {
	message := fmt.Sprintf("Changing the %s in an instance requires stopping it. ", strings.Join(propNames, ", "))
	if !d.Get("allow_stopping_for_update").(bool) {
		return fmt.Errorf(message + "To acknowledge this action, please set allow_stopping_for_update = true in your config file.")
	}
	return nil
}

func hostnameDiffSuppressFunc(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.TrimRight(oldValue, ".") == strings.TrimRight(newValue, ".")
}
