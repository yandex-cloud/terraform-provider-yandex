package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexServerlessContainer() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Serverless Container. This data source is used to define Yandex Cloud Container that can be used by other resources.\n\n~> Either `container_id` or `name` must be specified.\n",

		ReadContext: dataSourceYandexServerlessContainerRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"container_id": {
				Type:        schema.TypeString,
				Description: "Yandex Cloud Serverless Container ID used to define container.",
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

			"memory": {
				Type:        schema.TypeInt,
				Description: resourceYandexServerlessContainer().Schema["memory"].Description,
				Computed:    true,
			},

			"cores": {
				Type:        schema.TypeInt,
				Description: resourceYandexServerlessContainer().Schema["cores"].Description,
				Computed:    true,
			},

			"core_fraction": {
				Type:        schema.TypeInt,
				Description: resourceYandexServerlessContainer().Schema["core_fraction"].Description,
				Computed:    true,
			},

			"execution_timeout": {
				Type:        schema.TypeString,
				Description: resourceYandexServerlessContainer().Schema["execution_timeout"].Description,
				Computed:    true,
			},

			"concurrency": {
				Type:        schema.TypeInt,
				Description: resourceYandexServerlessContainer().Schema["concurrency"].Description,
				Computed:    true,
			},

			"service_account_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Computed:    true,
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
						"mount_point_path": {
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
						"mount_point_path": {
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

			"image": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"work_dir": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"command": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"args": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"environment": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},

			"url": {
				Type:        schema.TypeString,
				Description: resourceYandexServerlessContainer().Schema["url"].Description,
				Computed:    true,
			},

			"revision_id": {
				Type:        schema.TypeString,
				Description: resourceYandexServerlessContainer().Schema["revision_id"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"runtime": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"http", "task"}, true),
						},
					},
				},
				Required: false,
				Optional: true,
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

			"async_invocation": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexServerlessContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "container_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}
	containerID := d.Get("container_id").(string)

	if _, ok := d.GetOk("name"); ok {
		containerID, err = resolveObjectID(ctx, config, d, sdkresolvers.ContainerResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Yandex Cloud Serverless Container by name: %v", err)
		}
	}

	req := containers.GetContainerRequest{
		ContainerId: containerID,
	}

	container, err := config.sdk.Serverless().Containers().Container().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Container %q", d.Id())))
	}

	revision, err := resolveContainerLastRevision(ctx, config, containerID)
	if err != nil {
		return diag.Errorf("Failed to resolve last revision of data source Yandex Cloud Container: %s", err)
	}

	d.SetId(container.Id)
	d.Set("container_id", container.Id)
	if revision != nil {
		d.Set("storage_mounts", flattenRevisionStorageMounts(revision.StorageMounts)) // for backward compatibility
	}

	return diag.FromErr(flattenYandexServerlessContainer(d, container, revision, true))
}
