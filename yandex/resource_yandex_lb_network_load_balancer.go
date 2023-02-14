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
)

const yandexLBNetworkLoadBalancerDefaultTimeout = 5 * time.Minute

func resourceYandexLBNetworkLoadBalancer() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Optional: true,
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

			"region_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"type": {
				Type:         schema.TypeString,
				Default:      "external",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"external", "internal"}, false),
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"listener": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourceLBNetworkLoadBalancerListenerHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"target_port": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp"}, false),
						},
						"external_address_spec": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      resourceLBNetworkLoadBalancerExternalAddressHash,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ip_version": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "ipv4",
										ValidateFunc: validation.StringInSlice([]string{"ipv4", "ipv6"}, false),
									},
								},
							},
						},
						"internal_address_spec": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      resourceLBNetworkLoadBalancerInternalAddressHash,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ip_version": {
										Type:         schema.TypeString,
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
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourceLBNetworkLoadBalancerAttachedTargetGroupHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"healthcheck": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"interval": {
										Type:     schema.TypeInt,
										Default:  2,
										Optional: true,
									},
									"timeout": {
										Type:     schema.TypeInt,
										Default:  1,
										Optional: true,
									},
									"unhealthy_threshold": {
										Type:     schema.TypeInt,
										Default:  2,
										Optional: true,
									},
									"healthy_threshold": {
										Type:     schema.TypeInt,
										Default:  2,
										Optional: true,
									},
									"http_options": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"path": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"tcp_options": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:     schema.TypeInt,
													Required: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
