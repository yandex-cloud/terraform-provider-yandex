package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func dataSourceYandexComputeInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeInstanceRead,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cores": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"core_fraction": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"boot_disk": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_delete": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"initialize_params": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"image_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"network_interface": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipv6": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ipv6_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"nat": {
							Type:     schema.TypeBool,
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
			"secondary_disk": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_delete": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_id": {
							Type:     schema.TypeString,
							Computed: true,
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
							ForceNew: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceYandexComputeInstanceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	var instance *compute.Instance

	instanceID := d.Get("instance_id").(string)
	instance, err := config.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: instanceID,
		View:       compute.InstanceView_FULL,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("instance with ID %q", instanceID))
	}

	resources, err := flattenInstanceResources(instance)
	if err != nil {
		return err
	}

	bootDisk, err := flattenInstanceBootDisk(instance, config.sdk.Compute().Disk())
	if err != nil {
		return err
	}

	networkInterfaces, _, _, err := flattenInstanceNetworkInterfaces(instance)
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

	createdAt, err := getTimestamp(instance.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("instance_id", instance.Id)
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

	if err := d.Set("network_interface", networkInterfaces); err != nil {
		return err
	}

	if err := d.Set("secondary_disk", secondaryDisks); err != nil {
		return err
	}

	if err := d.Set("scheduling_policy", schedulingPolicy); err != nil {
		return err
	}

	d.SetId(instance.Id)

	return nil
}
