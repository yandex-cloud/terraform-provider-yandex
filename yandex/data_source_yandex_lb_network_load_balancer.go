package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexLBNetworkLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexLBNetworkLoadBalancerRead,
		Schema: map[string]*schema.Schema{
			"network_load_balancer_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region_id": {
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
			"listener": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      resourceLBNetworkLoadBalancerListenerHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"target_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_address_spec": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      resourceLBNetworkLoadBalancerExternalAddressHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"internal_address_spec": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      resourceLBNetworkLoadBalancerInternalAddressHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"attached_target_group": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      resourceLBNetworkLoadBalancerAttachedTargetGroupHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"healthcheck": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"interval": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"timeout": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"unhealthy_threshold": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"healthy_threshold": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"http_options": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"path": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"tcp_options": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:     schema.TypeInt,
													Computed: true,
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
				Computed: true,
			},
		},
	}
}

func dataSourceYandexLBNetworkLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "network_load_balancer_id", "name")
	if err != nil {
		return err
	}

	nlbID := d.Get("network_load_balancer_id").(string)
	_, nlbNameOk := d.GetOk("name")

	if nlbNameOk {
		nlbID, err = resolveObjectID(ctx, config, d, sdkresolvers.NetworkLoadBalancerResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source network load balancer by name: %v", err)
		}
	}

	nlb, err := config.sdk.LoadBalancer().NetworkLoadBalancer().Get(ctx, &loadbalancer.GetNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: nlbID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("network load balancer with ID %q", nlbID))
	}

	ls, err := flattenLBListenerSpecs(nlb)
	if err != nil {
		return err
	}

	atgs, err := flattenLBAttachedTargetGroups(nlb)
	if err != nil {
		return err
	}

	d.Set("network_load_balancer_id", nlb.Id)
	d.Set("name", nlb.Name)
	d.Set("type", strings.ToLower(nlb.Type.String()))
	d.Set("region_id", nlb.RegionId)
	d.Set("description", nlb.Description)
	d.Set("created_at", getTimestamp(nlb.CreatedAt))
	d.Set("folder_id", nlb.FolderId)
	d.Set("deletion_protection", nlb.DeletionProtection)

	if err := d.Set("labels", nlb.Labels); err != nil {
		return err
	}

	if err := d.Set("listener", ls); err != nil {
		return err
	}

	if err := d.Set("attached_target_group", atgs); err != nil {
		return err
	}

	d.SetId(nlb.Id)

	return nil
}
