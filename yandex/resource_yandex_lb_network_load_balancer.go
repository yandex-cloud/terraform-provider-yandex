package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexLBNetworkLoadBalancerDefaultTimeout = 5 * time.Minute

func resourceYandexLBNetworkLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a network load balancer in the specified folder using the data specified in the config. For more information, see [the official documentation](https://yandex.cloud/docs/load-balancer/concepts).",

		Create: resourceYandexLBNetworkLoadBalancerCreate,
		Read:   resourceYandexLBNetworkLoadBalancerRead,
		Update: resourceYandexLBNetworkLoadBalancerUpdate,
		Delete: resourceYandexLBNetworkLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexLBNetworkLoadBalancerDefaultTimeout),
			Update: schema.DefaultTimeout(yandexLBNetworkLoadBalancerDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexLBNetworkLoadBalancerDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"region_id": {
				Type:        schema.TypeString,
				Description: "ID of the availability zone where the network load balancer resides. If omitted, default region is being used.",
				Optional:    true,
				Computed:    true,
			},

			"type": {
				Type:         schema.TypeString,
				Description:  "Type of the network load balancer. Must be one of 'external' or 'internal'. The default is 'external'.",
				Default:      "external",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"external", "internal"}, false),
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"listener": {
				Type:        schema.TypeSet,
				Description: "Listener specification that will be used by a network load balancer.\n\n~> One of `external_address_spec` or `internal_address_spec` should be specified.\n",
				Optional:    true,
				Set:         resourceLBNetworkLoadBalancerListenerHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the listener. The name must be unique for each listener on a single load balancer.",
							Required:    true,
						},
						"port": {
							Type:        schema.TypeInt,
							Description: "Port for incoming traffic.",
							Required:    true,
						},
						"target_port": {
							Type:        schema.TypeInt,
							Description: "Port of a target. The default is the same as listener's port.",
							Optional:    true,
							Computed:    true,
						},
						"protocol": {
							Type:         schema.TypeString,
							Description:  "Protocol for incoming traffic. TCP or UDP and the default is TCP.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp"}, false),
						},
						"external_address_spec": {
							Type:        schema.TypeSet,
							Description: "External IP address specification. ",
							Optional:    true,
							Set:         resourceLBNetworkLoadBalancerExternalAddressHash,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:        schema.TypeString,
										Description: "External IP address for a listener. IP address will be allocated if it wasn't been set.",
										Optional:    true,
										Computed:    true,
									},
									"ip_version": {
										Type:         schema.TypeString,
										Description:  "IP version of the external addresses that the load balancer works with. Must be one of `ipv4` or `ipv6`. The default is `ipv4`.",
										Optional:     true,
										Default:      "ipv4",
										ValidateFunc: validation.StringInSlice([]string{"ipv4", "ipv6"}, false),
									},
								},
							},
						},
						"internal_address_spec": {
							Type:        schema.TypeSet,
							Description: "Internal IP address specification. ",
							Optional:    true,
							Set:         resourceLBNetworkLoadBalancerInternalAddressHash,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:        schema.TypeString,
										Description: "ID of the subnet to which the internal IP address belongs.",
										Required:    true,
									},
									"address": {
										Type:        schema.TypeString,
										Description: "Internal IP address for a listener. Must belong to the subnet that is referenced in subnet_id. IP address will be allocated if it wasn't been set.",
										Optional:    true,
										Computed:    true,
									},
									"ip_version": {
										Type:         schema.TypeString,
										Description:  "IP version of the external addresses that the load balancer works with. Must be one of `ipv4` or `ipv6`. The default is `ipv4`.",
										Optional:     true,
										Default:      "ipv4",
										ValidateFunc: validation.StringInSlice([]string{"ipv4", "ipv6"}, false),
									},
								},
							},
						},
					},
				},
			},

			"attached_target_group": {
				Type:        schema.TypeSet,
				Description: "An AttachedTargetGroup resource.",
				Optional:    true,
				Set:         resourceLBNetworkLoadBalancerAttachedTargetGroupHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_id": {
							Type:        schema.TypeString,
							Description: "ID of the target group.",
							Required:    true,
						},
						"healthcheck": {
							Type:        schema.TypeList,
							Description: "A HealthCheck resource.\n\n~> One of `http_options` or `tcp_options` should be specified.\n",
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Description: "Name of the health check. The name must be unique for each target group that attached to a single load balancer.",
										Required:    true,
									},
									"interval": {
										Type:        schema.TypeInt,
										Description: "The interval between health checks. The default is 2 seconds.",
										Default:     2,
										Optional:    true,
									},
									"timeout": {
										Type:        schema.TypeInt,
										Description: "Timeout for a target to return a response for the health check. The default is 1 second.",
										Default:     1,
										Optional:    true,
									},
									"unhealthy_threshold": {
										Type:        schema.TypeInt,
										Description: "Number of failed health checks before changing the status to `UNHEALTHY`. The default is 2.",
										Default:     2,
										Optional:    true,
									},
									"healthy_threshold": {
										Type:        schema.TypeInt,
										Description: "Number of successful health checks required in order to set the `HEALTHY` status for the target.",
										Default:     2,
										Optional:    true,
									},
									"http_options": {
										Type:        schema.TypeList,
										Description: "Options for HTTP health check.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:        schema.TypeInt,
													Description: "Port to use for HTTP health checks.",
													Required:    true,
												},
												"path": {
													Type:        schema.TypeString,
													Description: "URL path to set for health checking requests for every target in the target group. For example `/ping`. The default path is `/`.",
													Optional:    true,
												},
											},
										},
									},
									"tcp_options": {
										Type:        schema.TypeList,
										Description: "Options for TCP health check.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:        schema.TypeInt,
													Description: "Port to use for TCP health checks.",
													Required:    true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"allow_zonal_shift": {
				Type:        schema.TypeBool,
				Description: "Flag that marks the network load balancer as available to zonal shift.",
				Optional:    true,
				Computed:    true,
			},
		},
	}

}

func resourceYandexLBNetworkLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating network load balancer: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating network load balancer: %s", err)
	}

	ls, err := expandLBListenerSpecs(d)
	if err != nil {
		return fmt.Errorf("Error expanding listeners while creating network load balancer: %s", err)
	}

	atgs, err := expandLBAttachedTargetGroups(d)
	if err != nil {
		return fmt.Errorf("Error expanding attached target groups while creating network load balancer: %s", err)
	}

	nlbType, err := parseNetworkLoadBalancerType(d.Get("type").(string))
	if err != nil {
		return fmt.Errorf("Error expanding balancer type while creating network load balancer: %s", err)
	}

	req := loadbalancer.CreateNetworkLoadBalancerRequest{
		FolderId:             folderID,
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		RegionId:             d.Get("region_id").(string),
		Type:                 nlbType,
		Labels:               labels,
		ListenerSpecs:        ls,
		AttachedTargetGroups: atgs,
		DeletionProtection:   d.Get("deletion_protection").(bool),
		AllowZonalShift:      d.Get("allow_zonal_shift").(bool),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.LoadBalancer().NetworkLoadBalancer().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create network load balancer: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get network load balancer create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*loadbalancer.CreateNetworkLoadBalancerMetadata)
	if !ok {
		return fmt.Errorf("could not get NetworkLoadBalancer ID from create operation metadata")
	}

	d.SetId(md.NetworkLoadBalancerId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create network load balancer: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Network creation failed: %s", err)
	}

	return resourceYandexLBNetworkLoadBalancerRead(d, meta)
}

func resourceYandexLBNetworkLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	nlb, err := config.sdk.LoadBalancer().NetworkLoadBalancer().Get(ctx, &loadbalancer.GetNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("NetworkLoadBalancer %q", d.Get("name").(string)))
	}

	ls, err := flattenLBListenerSpecs(nlb)
	if err != nil {
		return err
	}

	atgs, err := flattenLBAttachedTargetGroups(nlb)
	if err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(nlb.CreatedAt))
	d.Set("name", nlb.Name)
	d.Set("folder_id", nlb.FolderId)
	d.Set("description", nlb.Description)
	d.Set("region_id", nlb.RegionId)
	d.Set("type", strings.ToLower(nlb.Type.String()))
	d.Set("deletion_protection", nlb.DeletionProtection)
	d.Set("allow_zonal_shift", nlb.AllowZonalShift)

	if err := d.Set("listener", ls); err != nil {
		return err
	}

	if err := d.Set("attached_target_group", atgs); err != nil {
		return err
	}

	return d.Set("labels", nlb.Labels)
}

func resourceYandexLBNetworkLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	ls, err := expandLBListenerSpecs(d)
	if err != nil {
		return fmt.Errorf("Error expanding listeners while creating network load balancer: %s", err)
	}

	atgs, err := expandLBAttachedTargetGroups(d)
	if err != nil {
		return fmt.Errorf("Error expanding attached target groups while creating network load balancer: %s", err)
	}

	req := &loadbalancer.UpdateNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: d.Id(),
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		Labels:                labels,
		ListenerSpecs:         ls,
		AttachedTargetGroups:  atgs,
		DeletionProtection:    d.Get("deletion_protection").(bool),
		AllowZonalShift:       d.Get("allow_zonal_shift").(bool),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.LoadBalancer().NetworkLoadBalancer().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update NetworkLoadBalancer %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating NetworkLoadBalancer %q: %s", d.Id(), err)
	}

	return resourceYandexLBNetworkLoadBalancerRead(d, meta)
}

func resourceYandexLBNetworkLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting NetworkLoadBalancer %q", d.Id())

	req := &loadbalancer.DeleteNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.LoadBalancer().NetworkLoadBalancer().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("NetworkLoadBalancer %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting NetworkLoadBalancer %q", d.Id())
	return nil
}
