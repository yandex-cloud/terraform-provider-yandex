package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexALBLoadBalancerDefaultTimeout = 10 * time.Minute

const (
	resourceNameRedirects   = "redirects"
	resourceNameHTTPToHTTPS = "http_to_https"
)

type redirect struct {
	httpToHTTPS bool
}

func resourceYandexALBLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description: "Creates an Application Load Balancer in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/application-load-balancer/concepts/application-load-balancer).",
		Create:      resourceYandexALBLoadBalancerCreate,
		Read:        resourceYandexALBLoadBalancerRead,
		Update:      resourceYandexALBLoadBalancerUpdate,
		Delete:      resourceYandexALBLoadBalancerDelete,
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
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.ResourceDescriptions["name"],
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.ResourceDescriptions["description"],
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: common.ResourceDescriptions["folder_id"],
			},

			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: common.ResourceDescriptions["labels"],
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: common.ResourceDescriptions["created_at"],
			},

			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The region ID where Load Balancer is located at.",
			},

			"network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: common.ResourceDescriptions["network_id"],
			},

			"log_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cloud Logging group ID to send logs to. Leave empty to use the balancer folder default log group.",
			},

			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the Load Balancer.",
			},

			"security_group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: common.ResourceDescriptions["security_group_ids"],
			},

			"allocation_policy": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Allocation zones for the Load Balancer instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Unique set of locations.",
							Set:         resourceALBAllocationPolicyLocationHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: common.ResourceDescriptions["zone"],
									},
									"subnet_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "ID of the subnet that location is located at.",
									},
									"disable_traffic": {
										Type:        schema.TypeBool,
										Default:     false,
										Optional:    true,
										Description: "If set, will disable all L7 instances in the zone for request handling.",
									},
								},
							},
						},
					},
				},
			},

			"log_options": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Cloud Logging settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Set to `true` to disable Cloud Logging for the balancer.",
						},
						"discard_rule": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of rules to discard a fraction of logs.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"discard_percent": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 100),
										Description:  "The percent of logs which will be discarded.",
									},

									"grpc_codes": {
										Type:        schema.TypeList,
										Description: "list of grpc codes by name, e.g, [**NOT_FOUND**, **RESOURCE_EXHAUSTED**].",
										Optional:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"http_code_intervals": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of http code intervals *1XX*-*5XX* or *ALL*",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"http_codes": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of http codes *100*-*599*.",
										Elem: &schema.Schema{
											Type:         schema.TypeInt,
											ValidateFunc: validation.IntBetween(100, 599),
										},
									},
								},
							},
						},

						"log_group_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile("^(([a-zA-Z][-a-zA-Z0-9_.]{0,63})?)$"), ""),
							Description:  "Cloud Logging group ID to send logs to. Leave empty to use the balancer folder default log group.",
						},
					},
				},
			},

			"listener": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of listeners for the Load Balancer.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the listener.",
						},
						"endpoint": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Network endpoint (addresses and ports) of the listener.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ports": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "One or more ports to listen on.",
										Elem:        &schema.Schema{Type: schema.TypeInt},
									},
									"address": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "One or more addresses to listen on.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"external_ipv4_address": {
													Type:        schema.TypeList,
													MaxItems:    1,
													Optional:    true,
													Description: "External IPv4 address.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:        schema.TypeString,
																Computed:    true,
																Optional:    true,
																Description: "Provided by the client or computed automatically.",
															},
														},
													},
												},
												"internal_ipv4_address": {
													Type:        schema.TypeList,
													MaxItems:    1,
													Optional:    true,
													Description: "Internal IPv4 address.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:        schema.TypeString,
																Computed:    true,
																Optional:    true,
																Description: "Provided by the client or computed automatically.",
															},
															"subnet_id": {
																Type:        schema.TypeString,
																Computed:    true,
																Optional:    true,
																Description: "Provided by the client or computed automatically.",
															},
														},
													},
												},
												"external_ipv6_address": {
													Type:        schema.TypeList,
													MaxItems:    1,
													Optional:    true,
													Description: "External IPv6 address.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:        schema.TypeString,
																Computed:    true,
																Optional:    true,
																Description: "Provided by the client or computed automatically.",
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
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "HTTP handler that sets plain text HTTP router.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": httpHandler(),
									"redirects": {
										Type:             schema.TypeList,
										MaxItems:         1,
										Optional:         true,
										DiffSuppressFunc: redirectsDiffSuppress,
										Description:      "Shortcut for adding http -> https redirects.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_to_https": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "If set redirects all unencrypted HTTP requests to the same URI with scheme changed to `https`.",
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
		},
	}
}

func tlsHandler() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Required:    true,
		Description: "TLS handler resource.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_handler":   httpHandler(),
				"stream_handler": streamHandler(),
				"certificate_ids": {
					Type:        schema.TypeSet,
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Set:         schema.HashString,
					Description: "Certificate IDs in the Certificate Manager. Multiple TLS certificates can be associated with the same context to allow both RSA and ECDSA certificates. Only the first certificate of each type will be used.",
				},
			},
		},
	}
}

func httpHandler() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Description: "Stream handler that sets plaintext Stream backend group.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_router_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "HTTP router id.",
				},
				"http2_options": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "If set, will enable HTTP2 protocol for the handler.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_concurrent_streams": {
								Type:        schema.TypeInt,
								Optional:    true,
								Default:     0,
								Description: "Maximum number of concurrent streams.",
							},
						},
					},
				},
				"allow_http10": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "If set, will enable only HTTP1 protocol with HTTP1.0 support.",
				},
				"rewrite_request_id": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "When unset, will preserve the incoming `x-request-id` header, otherwise would rewrite it with a new value.",
				},
			},
		},
	}
}

func streamHandler() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Description: "Stream handler resource.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backend_group_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Backend Group ID.",
				},
				"idle_timeout": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The idle timeout is the interval after which the connection will be forcibly closed if no data has been transmitted or received on either the upstream or downstream connection. If not configured, the default idle timeout is 1 hour. Setting it to 0 disables the timeout.",
				},
			},
		},
	}
}

func buildALBLoadBalancerCreateRequest(d *schema.ResourceData, config *Config) (*apploadbalancer.CreateLoadBalancerRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating ALB Load Balancer: %w", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating ALB Load Balancer: %w", err)
	}

	req := &apploadbalancer.CreateLoadBalancerRequest{
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
		return nil, fmt.Errorf("Error expanding allocation policy while creating ALB Load Balancer: %w", err)
	}
	req.SetAllocationPolicy(allocationPolicy)

	logOptions, err := expandALBLogOptions(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding log options while creating ALB Load Balancer: %w", err)
	}
	req.SetLogOptions(logOptions)

	listeners, err := expandALBListeners(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding listeners while creating ALB Load Balancer: %w", err)
	}
	req.SetListenerSpecs(listeners)

	return req, nil
}

func resourceYandexALBLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating ALB Load Balancer %q", d.Id())

	config := meta.(*Config)

	req, err := buildALBLoadBalancerCreateRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().LoadBalancer().Create(ctx, req))
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

	logOptions, err := flattenALBLogOptions(alb)
	if err != nil {
		return err
	}
	if err = d.Set("log_options", logOptions); err != nil {
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

	logOptions, err := expandALBLogOptions(d)
	if err != nil {
		return err
	}
	req.SetLogOptions(logOptions)

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

// redirectsDiffSuppress is a custom diff function for http redirects resource and it's inner fields.
//
// Main thing is to suppress diff between nil and empty redirect objects since they have no sense and we
// do not handle them during create or update operations.
//
// Handles redirect lists with at most 1 element, panics if any state contains more than 1 element.
func redirectsDiffSuppress(key, oldValue, newValue string, d *schema.ResourceData) bool {
	if strings.HasSuffix(key, resourceNameRedirects+".#") {
		var oldRedirectUntyped, newRedirectUntyped interface{}

		oldRedirectsUntyped, newRedirectsUntyped := d.GetChange(strings.ReplaceAll(key, ".#", ""))

		oldRedirects := oldRedirectsUntyped.([]interface{})
		newRedirects := newRedirectsUntyped.([]interface{})

		if len(oldRedirects) > 1 {
			panic("redirects diff suppress: too many redirect elements for previous state")
		}

		if len(newRedirects) > 1 {
			panic("redirects diff suppress: too many redirect elements for new state")
		}

		if len(oldRedirects) == 1 {
			oldRedirectUntyped = oldRedirects[0]
		}

		if len(newRedirects) == 1 {
			newRedirectUntyped = newRedirects[0]
		}

		oldRedirect := expandRedirect(oldRedirectUntyped)
		newRedirect := expandRedirect(newRedirectUntyped)

		return oldRedirect == newRedirect
	}

	if !strings.HasSuffix(key, resourceNameHTTPToHTTPS) {
		panic(fmt.Sprintf("redirects diff suppress: unexpected resource key '%v'", key))
	}

	return oldValue == newValue
}

// expandRedirect parses redirect object from dynamic data.
//
// Panics on any type assertion error.
func expandRedirect(data interface{}) redirect {
	r := redirect{}

	if data == nil {
		return r
	}

	redirectMap := data.(map[string]interface{})

	if v, ok := redirectMap[resourceNameHTTPToHTTPS]; ok {
		r.httpToHTTPS = v.(bool)
	}

	return r
}
