package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexALBBackendGroupDefaultTimeout = 5 * time.Minute

const (
	expectedStatusesSchemaKey                   = "expected_statuses"
	keepConnectionsOnHostHealthFailureSchemaKey = "keep_connections_on_host_health_failure"
)

func resourceYandexALBBackendGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a backend group in the specified folder and adds the specified backends to it. For more information, see [the official documentation](https://yandex.cloud/docs/application-load-balancer/concepts/backend-group).\n\n~> Only one type of backends `http_backend` or `grpc_backend` or `stream_backend` should be specified.\n",
		Create:      resourceYandexALBBackendGroupCreate,
		Read:        resourceYandexALBBackendGroupRead,
		Update:      resourceYandexALBBackendGroupUpdate,
		Delete:      resourceYandexALBBackendGroupDelete,
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

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"session_affinity": sessionAffinity(),

			"http_backend": {
				Type:          schema.TypeList,
				Description:   "HTTP backend specification that will be used by the ALB Backend Group.\n\n~> Only one of `target_group_ids` or `storage_bucket` should be specified.\n",
				Optional:      true,
				ConflictsWith: []string{"grpc_backend", "stream_backend"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the backend.",
							Required:    true,
						},
						"weight": {
							Type:        schema.TypeInt,
							Description: "Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.",
							Optional:    true,
							Default:     1,
						},
						"port": {
							Type:         schema.TypeInt,
							Description:  "Port for incoming traffic.",
							ValidateFunc: validation.IntBetween(0, 65535),
							Optional:     true,
						},
						"load_balancing_config": loadBalancingConfig(),
						"healthcheck":           healthCheck(),
						"tls":                   tlsBackend(),
						// List of ID's of target groups that belong to the backend.
						"target_group_ids": {
							Type:        schema.TypeList,
							Description: "References target groups for the backend.",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						// A resource for Object Storage bucket used as a backend
						"storage_bucket": {
							Type:        schema.TypeString,
							Description: "Name of bucket which should be used as a backend.",
							Optional:    true,
						},
						"http2": {
							Type:        schema.TypeBool,
							Description: "Enables HTTP2 for upstream requests. If not set, HTTP 1.1 will be used by default.",
							Optional:    true,
						},
					},
				},
			},
			"stream_backend": {
				Type:          schema.TypeList,
				Description:   "Stream backend specification that will be used by the ALB Backend Group.",
				Optional:      true,
				ConflictsWith: []string{"grpc_backend", "http_backend"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the backend.",
							Required:    true,
						},
						"weight": {
							Type:        schema.TypeInt,
							Description: "Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.",
							Optional:    true,
							Default:     1,
						},
						"port": {
							Type:         schema.TypeInt,
							Description:  "Port for incoming traffic.",
							ValidateFunc: validation.IntBetween(0, 65535),
							Optional:     true,
						},
						"load_balancing_config": loadBalancingConfig(),
						"healthcheck":           healthCheck(),
						"tls":                   tlsBackend(),
						"target_group_ids": {
							Type:        schema.TypeList,
							Description: "References target groups for the backend.",
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"enable_proxy_protocol": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						keepConnectionsOnHostHealthFailureSchemaKey: {
							Type:        schema.TypeBool,
							Description: "If set, when a backend host becomes unhealthy (as determined by the configured health checks), keep connections to the failed host.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"grpc_backend": {
				Type:          schema.TypeList,
				Description:   "gRPC backend specification that will be used by the ALB Backend Group.",
				Optional:      true,
				ConflictsWith: []string{"http_backend", "stream_backend"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the backend.",
							Required:    true,
						},
						"weight": {
							Type:        schema.TypeInt,
							Description: "Weight of the backend. Traffic will be split between backends of the same BackendGroup according to their weights.",
							Optional:    true,
							Default:     1,
						},
						"port": {
							Type:         schema.TypeInt,
							Description:  "Port for incoming traffic.",
							ValidateFunc: validation.IntBetween(0, 65535),
							Optional:     true,
						},
						"load_balancing_config": loadBalancingConfig(),
						"healthcheck":           healthCheck(),
						"tls":                   tlsBackend(),
						"target_group_ids": {
							Type:        schema.TypeList,
							Description: "References target groups for the backend.",
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
		},
	}
}

func sessionAffinity() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Session affinity mode determines how incoming requests are grouped into one session.\n\n~> Only one type(`connection` or `cookie` or `header`) of session affinity should be specified.\n",
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"connection": {
					Type:        schema.TypeList,
					Description: "Requests received from the same IP are combined into a session. Stream backend groups only support session affinity by client IP address.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"source_ip": {
								Type:        schema.TypeBool,
								Description: "Source IP address to use with affinity.",
								Optional:    true,
							},
						},
					},
					Optional: true,
				},

				"cookie": {
					Type:        schema.TypeList,
					Description: "Requests with the same cookie value and the specified file name are combined into a session. Allowed only for `HTTP` and `gRPC` backend groups.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:         schema.TypeString,
								Description:  "Name of the HTTP cookie to use with affinity.",
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},

							"ttl": {
								Type:             schema.TypeString,
								Description:      "TTL for the cookie (if not set, session cookie will be used).",
								Optional:         true,
								DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
							},
						},
					},
					Optional: true,
				},

				"header": {
					Type:        schema.TypeList,
					Description: "Requests with the same value of the specified HTTP header, such as with user authentication data, are combined into a session. Allowed only for `HTTP` and `gRPC` backend groups.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"header_name": {
								Type:         schema.TypeString,
								Description:  "The name of the request header that will be used with affinity.",
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
					Optional: true,
				},
			},
		},
	}
}

func loadBalancingConfig() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Load Balancing Config specification that will be used by this backend.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"panic_threshold": {
					Type:         schema.TypeInt,
					Description:  "If percentage of healthy hosts in the backend is lower than panic_threshold, traffic will be routed to all backends no matter what the health status is. This helps to avoid healthy backends overloading when everything is bad. Zero means no panic threshold.",
					ValidateFunc: validation.IntBetween(0, 100),
					Optional:     true,
				},
				"locality_aware_routing_percent": {
					Type:         schema.TypeInt,
					Description:  "Percent of traffic to be sent to the same availability zone. The rest will be equally divided between other zones.",
					ValidateFunc: validation.IntBetween(0, 100),
					Optional:     true,
				},
				"strict_locality": {
					Type:        schema.TypeBool,
					Description: "If set, will route requests only to the same availability zone. Balancer won't know about endpoints in other zones.",
					Optional:    true,
				},
				"mode": {
					Type:         schema.TypeString,
					Description:  "Load balancing mode for the backend. Possible values: `ROUND_ROBIN`, `RANDOM`, `LEAST_REQUEST`, `MAGLEV_HASH`.",
					Optional:     true,
					Default:      "ROUND_ROBIN",
					ValidateFunc: validation.StringInSlice([]string{"ROUND_ROBIN", "RANDOM", "LEAST_REQUEST", "MAGLEV_HASH"}, false),
				},
			},
		},
	}
}

func healthCheck() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Healthcheck specification that will be used by this backend.\n\n~> Only one of `stream_healthcheck` or `http_healthcheck` or `grpc_healthcheck` should be specified.\n",
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"timeout": {
					Type:        schema.TypeString,
					Description: "Time to wait for a health check response.",
					Required:    true,
				},
				"interval": {
					Type:        schema.TypeString,
					Description: "Interval between health checks.",
					Required:    true,
				},
				"interval_jitter_percent": {
					Type:        schema.TypeFloat,
					Description: "An optional jitter amount as a percentage of interval. If specified, during every interval value of (interval_ms * interval_jitter_percent / 100) will be added to the wait time.",
					Optional:    true,
				},
				"healthy_threshold": {
					Type:        schema.TypeInt,
					Description: "Number of consecutive successful health checks required to promote endpoint into the healthy state. 0 means 1. Note that during startup, only a single successful health check is required to mark a host healthy.",
					Optional:    true,
				},
				"unhealthy_threshold": {
					Type:        schema.TypeInt,
					Description: "Number of consecutive failed health checks required to demote endpoint into the unhealthy state. 0 means 1. Note that for HTTP health checks, a single 503 immediately makes endpoint unhealthy.",
					Optional:    true,
				},
				"healthcheck_port": {
					Type:         schema.TypeInt,
					Description:  "Optional alternative port for health checking.",
					ValidateFunc: validation.IntBetween(0, 65535),
					Optional:     true,
				},
				"stream_healthcheck": {
					Type:        schema.TypeList,
					Description: "Stream Healthcheck specification that will be used by this healthcheck.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"send": {
								Type:        schema.TypeString,
								Description: "Message sent to targets during TCP data transfer. If not specified, no data is sent to the target.",
								Optional:    true,
							},
							"receive": {
								Type:        schema.TypeString,
								Description: "Data that must be contained in the messages received from targets for a successful health check. If not specified, no messages are expected from targets, and those that are received are not checked.",
								Optional:    true,
							},
						},
					},
				},
				"http_healthcheck": {
					Type:        schema.TypeList,
					Description: "HTTP Healthcheck specification that will be used by this healthcheck.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:        schema.TypeString,
								Description: "`Host` HTTP header value.",
								Optional:    true,
							},
							"path": {
								Type:        schema.TypeString,
								Description: "HTTP path.",
								Required:    true,
							},
							"http2": {
								Type:        schema.TypeBool,
								Description: "If set, health checks will use HTTP2.",
								Optional:    true,
							},
							expectedStatusesSchemaKey: {
								Type:        schema.TypeList,
								Description: "A list of HTTP response statuses considered healthy.",
								Elem: &schema.Schema{
									Type:         schema.TypeInt,
									ValidateFunc: validation.IntBetween(100, 599),
								},
								Optional: true,
							},
						},
					},
				},
				"grpc_healthcheck": {
					Type:        schema.TypeList,
					Description: "gRPC Healthcheck specification that will be used by this healthcheck.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"service_name": {
								Type:        schema.TypeString,
								Description: "Service name for `grpc.health.v1.HealthCheckRequest` message.",
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

func tlsBackend() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "TLS specification that will be used by this backend.\n\n~> Only one of `validation_context.0.trusted_ca_id` or `validation_context.0.trusted_ca_bytes` should be specified.\n",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"sni": {
					Type:        schema.TypeString,
					Description: "[SNI](https://en.wikipedia.org/wiki/Server_Name_Indication) string for TLS connections.",
					Optional:    true,
				},
				"validation_context": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"trusted_ca_id": {
								Type:        schema.TypeString,
								Description: "Trusted CA certificate ID in the Certificate Manager.",
								Optional:    true,
							},
							"trusted_ca_bytes": {
								Type:        schema.TypeString,
								Description: "PEM-encoded trusted CA certificate chain.",
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

func buildALBBackendGroupCreateRequest(d *schema.ResourceData, folderID string) (*apploadbalancer.CreateBackendGroupRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating Application Backend Group: %w", err)
	}

	req := &apploadbalancer.CreateBackendGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	_, ok := d.GetOk("http_backend")
	if ok {
		backend, err := expandALBHTTPBackends(d)
		if err != nil {
			return nil, fmt.Errorf("Error expanding http backends while creating Application Backend Group: %w", err)
		}
		req.SetHttp(backend)
	}

	_, ok = d.GetOk("grpc_backend")
	if ok {
		backend, err := expandALBGRPCBackends(d)
		if err != nil {
			return nil, fmt.Errorf("Error expanding grpc backends while creating Application Backend Group: %w", err)
		}
		req.SetGrpc(backend)
	}

	_, ok = d.GetOk("stream_backend")
	if ok {
		backend, err := expandALBStreamBackends(d)
		if err != nil {
			return nil, fmt.Errorf("Error expanding stream backends while creating Application Backend Group: %w", err)
		}
		req.SetStream(backend)
	}

	return req, nil
}

func resourceYandexALBBackendGroupCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating Application Backend Group %q", d.Id())

	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Application Backend Group: %w", err)
	}

	req, err := buildALBBackendGroupCreateRequest(d, folderID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().BackendGroup().Create(ctx, req))
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
		affinity, err := flattenALBHTTPSessionAffinity(bg.GetHttp())
		if err != nil {
			return err
		}
		if err := d.Set("session_affinity", affinity); err != nil {
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
		affinity, err := flattenALBGRPCSessionAffinity(bg.GetGrpc())
		if err != nil {
			return err
		}
		if err := d.Set("session_affinity", affinity); err != nil {
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
		affinity, err := flattenALBStreamSessionAffinity(bg.GetStream())
		if err != nil {
			return err
		}
		if err := d.Set("session_affinity", affinity); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Finished reading Application Backend Group %q", d.Id())
	return d.Set("labels", bg.Labels)
}

func buildALBBackendGroupUpdateRequest(d *schema.ResourceData) (*apploadbalancer.UpdateBackendGroupRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("Error expanding http backends while creating Application Backend Group: %w", err)
		}
		req.SetHttp(backend)
	}
	_, ok = d.GetOk("grpc_backend")
	if ok {
		backend, err := expandALBGRPCBackends(d)
		if err != nil {
			return nil, fmt.Errorf("Error expanding grpc backends while creating Application Backend Group: %w", err)
		}
		req.SetGrpc(backend)
	}

	_, ok = d.GetOk("stream_backend")
	if ok {
		backend, err := expandALBStreamBackends(d)
		if err != nil {
			return nil, fmt.Errorf("Error expanding stream backends while creating Application Backend Group: %w", err)
		}
		req.SetStream(backend)
	}

	return req, nil
}

func resourceYandexALBBackendGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Application Backend Group %q", d.Id())
	config := meta.(*Config)

	req, err := buildALBBackendGroupUpdateRequest(d)
	if err != nil {
		return err
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
