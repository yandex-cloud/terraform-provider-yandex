package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexFunction() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexFunctionRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"function_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
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

			"runtime": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"entrypoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"execution_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"environment": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"secrets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"environment_variable": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"storage_mounts": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mount_point_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"read_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Deprecated: useResourceInstead("storage_mounts", "mounts"),
			},

			"mounts": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"rw", "ro"}, true),
						},
						"ephemeral_disk": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size_gb": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"block_size_kb": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"object_storage": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Required: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connectivity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"async_invocation": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retries_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ymq_success_target": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"service_account_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"ymq_failure_target": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"service_account_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"log_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"log_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"folder_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"min_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tmpfs_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"concurrency": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "function_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	functionID := d.Get("function_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		functionID, err = resolveObjectID(ctx, config, d, sdkresolvers.FunctionResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Yandex Cloud Function by name: %v", err)
		}
	}

	req := functions.GetFunctionRequest{
		FunctionId: functionID,
	}

	function, err := config.sdk.Serverless().Functions().Function().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %q", d.Id())))
	}

	version, err := resolveFunctionLatestVersion(ctx, config, function.GetId())
	if err != nil {
		return diag.Errorf("Failed to get latest version of Yandex Function: %s", err)
	}

	d.SetId(function.Id)
	d.Set("function_id", function.Id)
	if version != nil {
		d.Set("storage_mounts", flattenVersionStorageMounts(version.StorageMounts)) // for backward compatibility
	}
	return diag.FromErr(flattenYandexFunction(d, function, version, true))
}
