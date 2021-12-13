package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const yandexALBVirtualHostDefaultTimeout = 5 * time.Minute

func resourceYandexALBVirtualHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexALBVirtualHostCreate,
		Read:   resourceYandexALBVirtualHostRead,
		Update: resourceYandexALBVirtualHostUpdate,
		Delete: resourceYandexALBVirtualHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBVirtualHostDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBVirtualHostDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBVirtualHostDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"http_router_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"authority": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"modify_request_headers":  headerModification("modify_request_headers."),
			"modify_response_headers": headerModification("modify_response_headers."),
			"route": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"http_route": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_route_action": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"timeout": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"idle_timeout": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"prefix_rewrite": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"upgrade_types": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
												"host_rewrite": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"auto_host_rewrite": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"redirect_action": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replace_scheme": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"replace_host": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"replace_port": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"remove_query": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"response_code": {
													Type:     schema.TypeString,
													Default:  "moved_permanently",
													Optional: true,
												},
												"replace_path": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"replace_prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"direct_response_action": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(100, 599),
													Optional:     true,
												},
												"body": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"http_match": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_method": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
												"path": stringMatch(),
											},
										},
									},
								},
							},
						},
						"grpc_route": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grpc_match": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqmn": stringMatch(),
											},
										},
									},
									"grpc_route_action": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"max_timeout": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"idle_timeout": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"host_rewrite": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"auto_host_rewrite": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"grpc_status_response_action": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:     schema.TypeString,
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
		},
	}
}

func stringMatch() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exact": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"prefix": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func headerModification(path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"append": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"replace": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"remove": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
	}
}

func resourceYandexALBVirtualHostCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Creating Application Virtual Host %q", d.Get("name"))
	authority, err := expandALBStringListFromSchemaSet(d.Get("authority"))
	if err != nil {
		return fmt.Errorf("Error expanding authority while updating Application Virtual Host: %w", err)
	}

	routes, err := expandALBRoutes(d)
	if err != nil {
		return fmt.Errorf("Error expanding routes while updating Application Virtual Host: %w", err)
	}

	requestHeaders, err := expandALBHeaderModification(d, "modify_request_headers")
	if err != nil {
		return fmt.Errorf("Error expanding modify request headers while updating Application Virtual Host: %w", err)
	}

	responseHeaders, err := expandALBHeaderModification(d, "modify_response_headers")
	if err != nil {
		return fmt.Errorf("Error expanding modify response headers while updating Application Virtual Host: %w", err)
	}
	req := apploadbalancer.CreateVirtualHostRequest{
		HttpRouterId:          d.Get("http_router_id").(string),
		Name:                  d.Get("name").(string),
		Authority:             authority,
		Routes:                routes,
		ModifyResponseHeaders: responseHeaders,
		ModifyRequestHeaders:  requestHeaders,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().VirtualHost().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Application Virtual Host: %w", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Application Virtual Host create operation metadata: %w", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateVirtualHostMetadata)
	if !ok {
		return fmt.Errorf("could not get Application Virtual Host ID from create operation metadata")
	}

	d.SetId(md.HttpRouterId + "/" + md.VirtualHostName)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Application Virtual Host: %w", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Application Virtual Host creation failed: %w", err)
	}

	log.Printf("[DEBUG] Finished creating Application Virtual Host %q", d.Id())
	return resourceYandexALBVirtualHostRead(d, meta)
}

func resourceYandexALBVirtualHostRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Application Virtual Host %q", d.Id())
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	virtualHost, err := config.sdk.ApplicationLoadBalancer().VirtualHost().Get(ctx, &apploadbalancer.GetVirtualHostRequest{
		HttpRouterId:    d.Get("http_router_id").(string),
		VirtualHostName: d.Get("name").(string),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Virtual Host %q", d.Get("name").(string)))
	}

	requestHeaderModification, err := flattenALBHeaderModification(virtualHost.ModifyRequestHeaders)
	if err != nil {
		return err
	}

	responseHeaderModification, err := flattenALBHeaderModification(virtualHost.ModifyResponseHeaders)
	if err != nil {
		return err
	}

	routes, err := flattenALBRoutes(virtualHost.Routes)
	if err != nil {
		return err
	}

	d.Set("name", virtualHost.Name)
	d.Set("authority", virtualHost.Authority)

	if err := d.Set("modify_request_headers", requestHeaderModification); err != nil {
		return err
	}

	if err := d.Set("modify_response_headers", responseHeaderModification); err != nil {
		return err
	}

	if err := d.Set("route", routes); err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished reading Application Virtual Host %q", d.Id())
	return nil
}

func resourceYandexALBVirtualHostUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating Application Virtual Host %q", d.Id())

	authority, err := expandALBStringListFromSchemaSet(d.Get("authority"))
	if err != nil {
		return fmt.Errorf("Error expanding authority while updating Application Virtual Host: %w", err)
	}

	routes, err := expandALBRoutes(d)
	if err != nil {
		return fmt.Errorf("Error expanding routes while updating Application Virtual Host: %w", err)
	}

	requestHeaders, err := expandALBHeaderModification(d, "modify_request_headers")
	if err != nil {
		return fmt.Errorf("Error expanding modify request headers while updating Application Virtual Host: %w", err)
	}

	responseHeaders, err := expandALBHeaderModification(d, "modify_response_headers")
	if err != nil {
		return fmt.Errorf("Error expanding modify response headers while updating Application Virtual Host: %w", err)
	}

	req := &apploadbalancer.UpdateVirtualHostRequest{
		VirtualHostName:       d.Get("name").(string),
		HttpRouterId:          d.Get("http_router_id").(string),
		Authority:             authority,
		Routes:                routes,
		ModifyResponseHeaders: responseHeaders,
		ModifyRequestHeaders:  requestHeaders,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().VirtualHost().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Application Virtual Host %q: %w", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Application Virtual Host %q: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Application Virtual Host %q", d.Id())
	return resourceYandexALBVirtualHostRead(d, meta)
}

func resourceYandexALBVirtualHostDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Application Virtual Host %q", d.Id())

	req := &apploadbalancer.DeleteVirtualHostRequest{
		VirtualHostName: d.Get("name").(string),
		HttpRouterId:    d.Get("http_router_id").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().VirtualHost().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Virtual Host %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Application Virtual Host %q", d.Id())
	return nil
}
