package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexFunction() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Function. For more information about Yandex Cloud Functions, see [Yandex Cloud Functions](https://yandex.cloud/docs/functions).\nThis data source is used to define [Yandex Cloud Function](https://yandex.cloud/docs/functions/concepts/function) that can be used by other resources.\n\n~> Either `function_id` or `name` must be specified.\n",

		ReadContext: dataSourceYandexFunctionRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"function_id": {
				Type:        schema.TypeString,
				Description: "Yandex Cloud Function id used to define function.",
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"runtime": {
				Type:        schema.TypeString,
				Description: resourceYandexFunction().Schema["runtime"].Description,
				Computed:    true,
			},

			"entrypoint": {
				Type:        schema.TypeString,
				Description: resourceYandexFunction().Schema["entrypoint"].Description,
				Computed:    true,
			},

			"memory": {
				Type:        schema.TypeInt,
				Description: resourceYandexFunction().Schema["memory"].Description,
				Computed:    true,
			},

			"execution_timeout": {
				Type:        schema.TypeString,
				Description: resourceYandexFunction().Schema["execution_timeout"].Description,
				Computed:    true,
			},

			"service_account_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Computed:    true,
			},

			"environment": {
				Type:        schema.TypeMap,
				Description: resourceYandexFunction().Schema["environment"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"tags": {
				Type:        schema.TypeSet,
				Description: resourceYandexFunction().Schema["tags"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
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
				Type:        schema.TypeString,
				Description: resourceYandexFunction().Schema["version"].Description,
				Computed:    true,
			},

			"image_size": {
				Type:        schema.TypeInt,
				Description: resourceYandexFunction().Schema["image_size"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
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
				Type:        schema.TypeInt,
				Description: resourceYandexFunction().Schema["tmpfs_size"].Description,
				Computed:    true,
			},

			"concurrency": {
				Type:        schema.TypeInt,
				Description: resourceYandexFunction().Schema["concurrency"].Description,
				Optional:    true,
				Computed:    true,
			},

			"metadata_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gce_http_endpoint": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 2),
							Optional:     true,
							Computed:     true,
						},
						"aws_v1_http_endpoint": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 2),
							Optional:     true,
							Computed:     true,
						},
					},
				},
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
