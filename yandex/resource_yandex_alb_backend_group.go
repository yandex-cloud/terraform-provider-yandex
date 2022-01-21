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

const yandexALBBackendGroupDefaultTimeout = 5 * time.Minute

func resourceYandexALBBackendGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexALBBackendGroupCreate,
		Read:   resourceYandexALBBackendGroupRead,
		Update: resourceYandexALBBackendGroupUpdate,
		Delete: resourceYandexALBBackendGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBBackendGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBBackendGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBBackendGroupDefaultTimeout),
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

			"http_backend": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"grpc_backend", "stream_backend"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"port": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 65535),
							Optional:     true,
						},
						"load_balancing_config": loadBalancingConfig(),
						"healthcheck":           healthcheck(),
						"tls":                   tlsBackend(),
						"target_group_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"http2": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Set: resourceALBBackendGroupBackendHash,
			},
			"stream_backend": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"grpc_backend", "http_backend"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"port": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 65535),
							Optional:     true,
						},
						"load_balancing_config": loadBalancingConfig(),
						"healthcheck":           healthcheck(),
						"tls":                   tlsBackend(),
						"target_group_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				Set: resourceALBBackendGroupBackendHash,
			},
			"grpc_backend": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"http_backend", "stream_backend"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"port": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 65535),
							Optional:     true,
						},
						"load_balancing_config": loadBalancingConfig(),
						"healthcheck":           healthcheck(),
						"tls":                   tlsBackend(),
						"target_group_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				Set: resourceALBBackendGroupBackendHash,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func loadBalancingConfig() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"panic_threshold": {
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(0, 100),
					Optional:     true,
				},
				"locality_aware_routing_percent": {
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(0, 100),
					Optional:     true,
				},
				"strict_locality": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
	}
}

func healthcheck() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"timeout": {
					Type:     schema.TypeString,
					Required: true,
				},
				"interval": {
					Type:     schema.TypeString,
					Required: true,
				},
				"interval_jitter_percent": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"healthy_threshold": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"unhealthy_threshold": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"healthcheck_port": {
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(0, 65535),
					Optional:     true,
				},
				"stream_healthcheck": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"send": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"receive": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"http_healthcheck": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"path": {
								Type:     schema.TypeString,
								Required: true,
							},
							"http2": {
								Type:     schema.TypeBool,
								Optional: true,
							},
						},
					},
				},
				"grpc_healthcheck": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"service_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
		Set: resourceALBBackendGroupHealthcheckHash,
	}
}

func tlsBackend() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"sni": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"validation_context": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"trusted_ca_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"trusted_ca_bytes": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func resourceYandexALBBackendGroupCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating Application Backend Group %q", d.Id())

	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Application Backend Group: %w", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Application Backend Group: %w", err)
	}

	req := apploadbalancer.CreateBackendGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	_, ok := d.GetOk("http_backend")
	if ok {
		backend, err := expandALBHTTPBackends(d)
		if err != nil {
			return fmt.Errorf("Error expanding http backends while creating Application Backend Group: %w", err)
		}
		req.SetHttp(backend)
	}
	_, ok = d.GetOk("grpc_backend")
	if ok {
		backend, err := expandALBGRPCBackends(d)
		if err != nil {
			return fmt.Errorf("Error expanding grpc backends while creating Application Backend Group: %w", err)
		}
		req.SetGrpc(backend)
	}
	_, ok = d.GetOk("stream_backend")
	if ok {
		backend, err := expandALBStreamBackends(d)
		if err != nil {
			return fmt.Errorf("Error expanding stream backends while creating Application Backend Group: %w", err)
		}
		req.SetStream(backend)
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().BackendGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Application Backend Group: %w", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Application Backend Group create operation metadata: %w", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateBackendGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get Application Backend Group ID from create operation metadata")
	}

	d.SetId(md.BackendGroupId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Application Backend Group: %w", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Application Backend Group creation failed: %w", err)
	}

	log.Printf("[DEBUG] Finished creating Application Backend Group %q", d.Id())
	return resourceYandexALBBackendGroupRead(d, meta)
}

func resourceYandexALBBackendGroupRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Application Backend Group %q", d.Id())
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	bg, err := config.sdk.ApplicationLoadBalancer().BackendGroup().Get(ctx, &apploadbalancer.GetBackendGroupRequest{
		BackendGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Backend Group %q", d.Get("name").(string)))
	}

	_ = d.Set("created_at", getTimestamp(bg.CreatedAt))
	_ = d.Set("name", bg.Name)
	_ = d.Set("folder_id", bg.FolderId)
	_ = d.Set("description", bg.Description)

	switch bg.GetBackend().(type) {
	case *apploadbalancer.BackendGroup_Http:
		backends, err := flattenALBHTTPBackends(bg)
		if err != nil {
			return err
		}
		if err := d.Set("http_backend", backends); err != nil {
			return err
		}
	case *apploadbalancer.BackendGroup_Grpc:
		backends, err := flattenALBGRPCBackends(bg)
		if err != nil {
			return err
		}
		if err := d.Set("grpc_backend", backends); err != nil {
			return err
		}
	case *apploadbalancer.BackendGroup_Stream:
		backends, err := flattenALBStreamBackends(bg)
		if err != nil {
			return err
		}
		if err := d.Set("stream_backend", backends); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Finished reading Application Backend Group %q", d.Id())
	return d.Set("labels", bg.Labels)
}

func resourceYandexALBBackendGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Application Backend Group %q", d.Id())
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	req := &apploadbalancer.UpdateBackendGroupRequest{
		BackendGroupId: d.Id(),
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		Labels:         labels,
	}

	_, ok := d.GetOk("http_backend")
	if ok {
		backend, err := expandALBHTTPBackends(d)
		if err != nil {
			return fmt.Errorf("Error expanding http backends while creating Application Backend Group: %w", err)
		}
		req.SetHttp(backend)
	}
	_, ok = d.GetOk("grpc_backend")
	if ok {
		backend, err := expandALBGRPCBackends(d)
		if err != nil {
			return fmt.Errorf("Error expanding grpc backends while creating Application Backend Group: %w", err)
		}
		req.SetGrpc(backend)
	}

	_, ok = d.GetOk("stream_backend")
	if ok {
		_, err := expandALBGRPCBackends(d)
		if err != nil {
			return fmt.Errorf("Error expanding stream backends while creating Application Backend Group: %w", err)
		}
		//req.SetStream(backend)
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().BackendGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Application Backend Group %q: %w", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Application Backend Group %q: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Application Backend Group %q", d.Id())
	return resourceYandexALBBackendGroupRead(d, meta)
}

func resourceYandexALBBackendGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting Application Backend Group %q", d.Id())
	config := meta.(*Config)

	req := &apploadbalancer.DeleteBackendGroupRequest{
		BackendGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().BackendGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Backend Group %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Application Backend Group %q", d.Id())
	return nil
}
