package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const yandexALBLoadBalancerDefaultTimeout = 10 * time.Minute

func resourceYandexALBLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexALBLoadBalancerCreate,
		Read:   resourceYandexALBLoadBalancerRead,
		Update: resourceYandexALBLoadBalancerUpdate,
		Delete: resourceYandexALBLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
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

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"region_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"log_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"allocation_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"disable_traffic": {
										Type:     schema.TypeBool,
										Default:  false,
										Optional: true,
									},
								},
							},
							Set: resourceALBAllocationPolicyLocationHash,
						},
					},
				},
			},

			"listener": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"endpoint": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ports": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
									"address": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"external_ipv4_address": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
																Optional: true,
															},
														},
													},
												},
												"internal_ipv4_address": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
																Optional: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Computed: true,
																Optional: true,
															},
														},
													},
												},
												"external_ipv6_address": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
																Optional: true,
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
						"http": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": httpHandler(),
									"redirects": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_to_https": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
											},
										},
									},
								},
							},
						},
						"stream": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": streamHandler(),
								},
							},
						},
						"tls": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_handler": tlsHandler(),
									"sni_handler": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"server_names": {
													Type:     schema.TypeSet,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
												"handler": tlsHandler(),
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
		},
	}
}

func tlsHandler() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_handler":   httpHandler(),
				"stream_handler": streamHandler(),
				"certificate_ids": {
					Type:     schema.TypeSet,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Set:      schema.HashString,
				},
			},
		},
	}
}

func httpHandler() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_router_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"http2_options": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_concurrent_streams": {
								Type:     schema.TypeInt,
								Optional: true,
								Default:  0,
							},
						},
					},
				},
				"allow_http10": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
			},
		},
	}
}

func streamHandler() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backend_group_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func resourceYandexALBLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating ALB Load Balancer %q", d.Id())

	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return fmt.Errorf("Error expanding labels while creating ALB Load Balancer: %w", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating ALB Load Balancer: %w", err)
	}

	req := apploadbalancer.CreateLoadBalancerRequest{
		FolderId:         folderID,
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		RegionId:         d.Get("region_id").(string),
		NetworkId:        d.Get("network_id").(string),
		SecurityGroupIds: expandStringSet(d.Get("security_group_ids")),
		Labels:           labels,
	}

	allocationPolicy, err := expandALBAllocationPolicy(d)
	if err != nil {
		return fmt.Errorf("Error expanding allocation policy while creating ALB Load Balancer: %w", err)
	}
	req.SetAllocationPolicy(allocationPolicy)

	listeners, err := expandALBListeners(d)
	if err != nil {
		return fmt.Errorf("Error expanding listeners while creating ALB Load Balancer: %w", err)
	}
	req.SetListenerSpecs(listeners)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().LoadBalancer().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create ALB Load Balancer: %w", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get ALB Load Balancer create operation metadata: %w", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateLoadBalancerMetadata)
	if !ok {
		return fmt.Errorf("could not get ALB Load Balancer ID from create operation metadata")
	}

	d.SetId(md.LoadBalancerId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create ALB Load Balancer: %w", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("ALB Load Balancer creation failed: %w", err)
	}

	log.Printf("[DEBUG] Finished creating ALB Load Balancer %q", d.Id())
	return resourceYandexALBLoadBalancerRead(d, meta)

}

func resourceYandexALBLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading ALB Load Balancer %q", d.Id())
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	alb, err := config.sdk.ApplicationLoadBalancer().LoadBalancer().Get(ctx, &apploadbalancer.GetLoadBalancerRequest{
		LoadBalancerId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("ALB Load Balancer %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(alb.CreatedAt))
	d.Set("name", alb.Name)
	d.Set("folder_id", alb.FolderId)
	d.Set("description", alb.Description)
	d.Set("region_id", alb.RegionId)
	d.Set("network_id", alb.NetworkId)
	d.Set("security_group_ids", alb.SecurityGroupIds)
	d.Set("log_group_id", alb.LogGroupId)
	d.Set("status", strings.ToLower(alb.Status.String()))

	allocationPolicy, err := flattenALBAllocationPolicy(alb)
	if err != nil {
		return err
	}
	if err := d.Set("allocation_policy", allocationPolicy); err != nil {
		return err
	}

	listeners, err := flattenALBListeners(alb)
	if err != nil {
		return err
	}
	if err := d.Set("listener", listeners); err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished reading ALB Load Balancer %q", d.Id())
	return d.Set("labels", alb.Labels)
}

func resourceYandexALBLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating ALB Load Balancer %q", d.Id())
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	req := &apploadbalancer.UpdateLoadBalancerRequest{
		LoadBalancerId:   d.Id(),
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		SecurityGroupIds: expandStringSet(d.Get("security_group_ids")),
		Labels:           labels,
	}

	allocationPolicy, err := expandALBAllocationPolicy(d)
	if err != nil {
		return fmt.Errorf("Error expanding allocation policy while creating ALB Load Balancer: %w", err)
	}
	req.SetAllocationPolicy(allocationPolicy)

	listeners, err := expandALBListeners(d)
	if err != nil {
		return fmt.Errorf("Error expanding listeners while creating ALB Load Balancer: %w", err)
	}
	req.SetListenerSpecs(listeners)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().LoadBalancer().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update ALB Load Balancer %q: %w", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating ALB Load Balancer %q: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating ALB Load Balancer %q", d.Id())
	return resourceYandexALBLoadBalancerRead(d, meta)
}

func resourceYandexALBLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting ALB Load Balancer %q", d.Id())
	config := meta.(*Config)

	req := &apploadbalancer.DeleteLoadBalancerRequest{
		LoadBalancerId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().LoadBalancer().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("ALB Load Balancer %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting ALB Load Balancer %q", d.Id())
	return nil
}
